package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tasktracker "github.com/vashc/async_arch_course/services/task_tracker/internal"
)

func main() {
	// Create config, e.g. via environment variables
	config, err := tasktracker.NewConfig()
	if err != nil {
		log.Fatalf("task_tracker.NewConfig error: %s", err.Error())
	}

	// Initialize storage
	storage, err := tasktracker.NewStorage(config)
	if err != nil {
		log.Fatalf("task_tracker.NewStorage error: %s", err.Error())
	}

	// Initialize Rabbit client and connection
	client, err := tasktracker.NewClient(config)
	if err != nil {
		log.Fatalf("task_tracker.NewClient error: %s", err.Error())
	}
	defer client.Close()

	// Create new chi application service
	service := tasktracker.NewService(config, storage, client)

	// Instantiate routes
	service.InstantiateRoutes()

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	worker := tasktracker.NewWorker(config, storage, client)
	go func() {
		err = worker.Process(ctx, tasktracker.RabbitQueue)
		if err != nil {
			log.Fatalf("worker.Process error: %s", err.Error())
		}
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down")
		cancel()
		if err = service.Stop(); err != nil {
			log.Fatalf("service.Stop error: %s", err.Error())
		}
	}()

	log.Printf(
		"Service is starting at %s:%s\n",
		config.API.Host,
		config.API.Port,
	)

	// Start handling requests
	err = service.Start()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Panicf("service.Start error: %s", err.Error())
	}

	log.Println("Running cleanup")
}
