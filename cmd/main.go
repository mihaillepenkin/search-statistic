package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"search_statistics/internal/config"
	handler "search_statistics/internal/infra/http"
	"search_statistics/internal/infra/rabbitmq"
	"search_statistics/internal/infra/repository"
	"search_statistics/internal/usecase"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel1 := context.WithCancel(context.Background())
	cfg := config.Load()
	rep := repository.New()
	cons := rabbitmq.New(cfg.RabbitMQURL, cfg.QueueName, rep)
	defer cons.Close()

	serv := usecase.NewUsecase(rep)
	handler := handler.MakeHandler(serv)
	m := mux.NewRouter()
	m.HandleFunc("/top/{n}", handler.GetTopHandler).Methods("GET")
	m.HandleFunc("/stoplist", handler.AddInStopListHandler).Methods("POST")
	m.HandleFunc("/stoplist", handler.DeleteFromStopList).Methods("DELETE")
	m.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           m,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Print("Server error: " + err.Error())
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	log.Print("always ok")
	go func() {
		sig := <-sigCh
		log.Print("Shutdown signal received, signal ", sig.String())
		log.Print("Shutting down server...")
		shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatal("Server forced to shutdown: " + err.Error())
		}
		cancel1()
	}()

	if err := cons.Start(ctx); err != nil && err != context.Canceled {
		log.Fatal("failed to start rabbitmq consumer")
	}

	log.Print("Service stopped gracefully")
}