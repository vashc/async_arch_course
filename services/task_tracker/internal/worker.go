package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

func NewWorker(config *Config, storage *Storage, rabbitClient *RabbitClient) *Worker {
	return &Worker{
		config:  config,
		storage: storage,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		rabbitClient: rabbitClient,
	}
}

func (w *Worker) Process(ctx context.Context, queueName string) error {
	queue, err := w.rabbitClient.Listen(queueName)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-queue:
			err = w.processOne(&msg)
			if err != nil {
				log.Printf("task_tracker.processOne error: %s\n", err.Error())
			}
		}
	}
}

func (w *Worker) processOne(msg *amqp.Delivery) (err error) {
	switch msg.Type {
	case string(userCreatedEventType):
		userCreatedIn := new(UserCreatedIn)
		err = json.Unmarshal(msg.Body, &userCreatedIn)
		if err != nil {
			return err
		}

		user := &User{
			ID:       userCreatedIn.ID,
			Username: userCreatedIn.Username,
			Role:     userCreatedIn.Role,
		}

		err = w.storage.CreateUser(user)
		if err != nil {
			return err
		}
	default:
		return nil
	}

	return nil
}
