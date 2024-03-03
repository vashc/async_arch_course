package internal

import (
	"context"
	"log"
	"math/rand"
	"time"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gocraft/dbr/v2"
	"github.com/google/uuid"
)

func NewService(config *Config, storage *Storage, client *RabbitClient) *Service {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", config.API.Host, config.API.Port),
		ReadHeaderTimeout: time.Second * 5,
	}

	service := &Service{
		config:  config,
		server:  server,
		storage: storage,
		client:  client,
		Mux:     chi.NewRouter(),
	}

	service.server.Handler = service

	return service
}

func (s *Service) InstantiateRoutes() {
	s.Use(
		middleware.Timeout(5*time.Second),
		LogRequest,
		MiddlewareUserAuth,
		MiddlewareUserCtx(s.config),
	)

	s.Route("/task", func(router chi.Router) {
		router.Post("/create", s.createTaskHandler())
		router.Post(
			fmt.Sprintf("/{%s}/complete", requestParamTaskID),
			s.completeTaskHandler(),
		)
		router.Get("/get", s.getTasksHandler())
		router.Post("/assign", s.assignTasksHandler())
	})

	s.Get("/health", s.healthHandler())
}

func (s *Service) Start() error {
	return s.server.ListenAndServe()
}

func (s *Service) Stop() error {
	err := s.storage.Close()
	if err != nil {
		return err
	}

	err = s.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) createTaskHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		task := new(Task)

		err := BodyParser(w, r, task)
		if err != nil {
			code := http.StatusUnprocessableEntity
			http.Error(w, http.StatusText(code), code)
			return
		}

		task.Status = createdStatus
		task.AuthorID, _ = r.Context().Value(requestParamUserID).(uuid.UUID)

		// Get a worker for the task randomly
		users, err := s.storage.GetUsersByRole(workerRole)
		if err != nil {
			log.Printf("storage.GetUsersByRole: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		if len(users) == 0 {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		rand.Seed(time.Now().Unix())
		//nolint:gosec // It's ok for now
		task.AssigneeID = users[rand.Intn(len(users))].ID

		if err = s.storage.CreateTask(task); err != nil {
			log.Printf("storage.CreateTask: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Create exchange message in a queue
		taskCreated := TaskCreatedOut{
			Description: task.Description,
			Status:      task.Status,
			AssigneeID:  task.AssigneeID,
		}

		err = s.client.Publish("", taskCreatedEventType, taskCreated)
		if err != nil {
			log.Printf("client.Publish: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		resp, err := json.Marshal(TaskCreateResponse{ID: task.ID})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) completeTaskHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(chi.URLParam(r, requestParamTaskID))
		if err != nil {
			log.Printf("uuid.Parse: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		task, err := s.storage.GetTaskByID(taskID)
		switch {
		case errors.Is(err, dbr.ErrNotFound):
			code := http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		case err != nil:
			log.Printf("storage.GetTaskByID: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		err = s.storage.UpdateTaskStatus(taskID, completedStatus)
		if err != nil {
			log.Printf("storage.UpdateTaskStatus: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Create exchange message in a queue
		taskCompleted := TaskCompletedOut{
			AssigneeID: task.AssigneeID,
		}

		err = s.client.Publish("", taskCompletedEventType, taskCompleted)
		if err != nil {
			log.Printf("client.Publish: %s\n", err.Error())
		}

		resp, err := json.Marshal(Response{Status: http.StatusText(http.StatusOK)})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) getTasksHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(requestParamUserID).(uuid.UUID)

		user, err := s.storage.GetUserByID(userID)
		if err != nil {
			log.Printf("storage.GetUserByID: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		if user.Role != workerRole {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			return
		}

		tasks, err := s.storage.GetTasksByAssigneeID(userID)
		switch {
		case errors.Is(err, dbr.ErrNotFound):
			code := http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		case err != nil:
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		resp, err := json.Marshal(tasks)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) assignTasksHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(requestParamUserID).(uuid.UUID)

		user, err := s.storage.GetUserByID(userID)
		if err != nil {
			log.Printf("storage.GetUserByID: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		if user.Role != adminRole && user.Role != managerRole {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			return
		}

		tasks, err := s.storage.GetTasksByStatus(createdStatus)
		switch {
		case errors.Is(err, dbr.ErrNotFound):
			code := http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		case err != nil:
			log.Printf("storage.GetTasksByStatus: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		users, err := s.storage.GetUsersByRole(workerRole)
		if err != nil {
			log.Printf("storage.GetUsersByRole: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		if len(users) == 0 {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		rand.Seed(time.Now().Unix())
		for _, task := range tasks {
			//nolint:gosec // It's ok for now
			task.AssigneeID = users[rand.Intn(len(users))].ID

			err = s.storage.UpdateTaskAssignee(task.ID, task.AssigneeID)
			if err != nil {
				log.Printf("storage.UpdateTaskAssignee: %s\n", err.Error())
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			// Create exchange message in a queue
			taskAssigned := TaskAssignedOut{
				AssigneeID: task.AssigneeID,
			}

			err = s.client.Publish("", taskAssignedEventType, taskAssigned)
			if err != nil {
				log.Printf("client.Publish: %s\n", err.Error())
			}
		}

		resp, err := json.Marshal(Response{Status: http.StatusText(http.StatusOK)})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) healthHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := json.Marshal(Response{Status: http.StatusText(http.StatusOK)})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}
