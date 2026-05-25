package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	amqp "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5673/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	queries := []string{
		"iphone 15 pro",
		"samsung galaxy s24",
		"xiaomi redmi note",
		"носки белые",
		"кроссовки nike",
		"платье летнее",
		"наушники bluetooth",
	}

	log.Println("sending 1000000 messages...")
	for i := 0; i < 1000000; i++ {
		query := queries[rand.Intn(len(queries))]
		body := fmt.Sprintf(`{"query":"%s","timestamp_sec":%d}`, query, time.Now().Unix())

		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			},
		)
		failOnError(err, "failed to publish a message")

		if i%1000 == 0 {
			log.Printf("Sent %d messages", i)
		}
		time.Sleep(10 * time.Millisecond)
	}
	log.Println("done sending messages")
}