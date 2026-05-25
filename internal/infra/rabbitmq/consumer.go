package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"search_statistics/internal/domain"
	"search_statistics/internal/usecase"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Consumer struct {
	rabbitMQURL string
	queueName   string
	conn *amqp.Connection
	ch   *amqp.Channel
	startTime   time.Time
	slotMu      sync.Mutex

	currentSlot map[string]int
	r           usecase.Repository
}

func New(rabbitMQURL, queueName string, r usecase.Repository) *Consumer {
	queueName = strings.TrimSpace(queueName)
	if queueName == "" {
		queueName = "ssh_commands"
	}


	return &Consumer{
		rabbitMQURL:     strings.TrimSpace(rabbitMQURL),
		queueName:       queueName,
		startTime:       time.Now(),
		currentSlot: make(map[string]int),
		r:           r,
	}
}

func (c *Consumer) Start(ctx context.Context) error {

	conn, err := amqp.Dial(c.rabbitMQURL)
	if err != nil {
		return fmt.Errorf("rabbitmq dial failed: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("rabbitmq channel failed: %w", err)
	}
	c.ch = ch

	q, err := ch.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue declare failed: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue consume failed: %w", err)
	}

	log.Print("rabbitmq consumer started, queue ", q.Name)

	go c.startProcessingTime(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq delivery channel closed")
			}
			c.handleMessage(msg)
		}
	}
}

func (c *Consumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	var message domain.Message
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		log.Print("invalid task payload ", err.Error())
		_ = msg.Reject(false)
		return
	}

	err := c.executeMessage(message)
	if err != nil {
		_ = msg.Ack(false)
		return
	}
	_ = msg.Ack(false)
}

func (c *Consumer) executeMessage(message domain.Message) error {
	if (message.Query == "") {
		return fmt.Errorf("invalid query format")
	}
	c.currentSlot[message.Query]++
	return nil
}

func (c *Consumer) startProcessingTime(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("time processing goroutine stopped")
			return
		case <-ticker.C:
			c.slotMu.Lock()
			snapshot := make(map[string]int, len(c.currentSlot))
            for k, v := range c.currentSlot {
                snapshot[k] = v
            }
			c.currentSlot = make(map[string]int)
			c.slotMu.Unlock()
			c.r.Update(snapshot)
		}
	}
}