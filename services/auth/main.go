package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	auth "github.com/vashc/async_arch_course/services/auth/internal"
)

func main() {
	// Create config, e.g. via environment variables
	config, err := auth.NewConfig()
	if err != nil {
		log.Fatalf("auth.NewConfig error: %s", err.Error())
	}

	// Initialize storage
	storage, err := auth.NewStorage(config)
	if err != nil {
		log.Fatalf("auth.NewStorage error: %s", err.Error())
	}

	// Initialize Rabbit client and connection
	client, err := auth.NewClient(config)
	if err != nil {
		log.Fatalf("auth.NewClient error: %s", err.Error())
	}
	defer client.Close()

	// Create new chi application service
	service := auth.NewService(config, storage, client)

	// Instantiate routes
	service.InstantiateRoutes()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down")
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
