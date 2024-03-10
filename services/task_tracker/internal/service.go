package internal

import (
	"bytes"
	"context"
	"log"
	"math/rand"
	"regexp"
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

func NewService(
	config *Config,
	storage *Storage,
	client *RabbitClient,
	httpClient *http.Client,
) *Service {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", config.API.Host, config.API.Port),
		ReadHeaderTimeout: time.Second * 5,
	}

	service := &Service{
		config:     config,
		server:     server,
		storage:    storage,
		client:     client,
		httpClient: httpClient,
		Mux:        chi.NewRouter(),
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

//nolint:funlen // It should be separated
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

		re := regexp.MustCompile(`(?P<prefix>\[[a-zA-Z0-9]*\])(?P<title>.*)`)
		matches := re.FindAllStringSubmatch(task.Title, -1)
		for i := range matches {
			task.JiraID = matches[i][1]
			task.Title = matches[i][2]
		}

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

		taskCost := &TaskCost{
			TaskID: task.ID,
			//nolint:gosec // It's ok for now
			AssignCost: 10 + rand.Intn(11),
			//nolint:gosec // It's ok for now
			CompleteCost: 20 + rand.Intn(21),
		}

		if err = s.storage.CreateTaskCost(taskCost); err != nil {
			log.Printf("storage.CreateTaskCost: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Create exchange message in a queue
		taskAssigned := TaskAssignedOut{
			Amount:     taskCost.AssignCost,
			AssigneeID: task.AssigneeID,
		}

		taskAssignedBody, err := json.Marshal(taskAssigned)
		if err != nil {
			log.Printf("json.Marshal: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		//nolint:noctx // Context is not used as for now
		request, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf(
				"%s/validate/%s/event",
				s.config.SchemaRegistryHost,
				taskAssignedEventType,
			),
			bytes.NewReader(taskAssignedBody),
		)
		if err != nil {
			log.Printf("http.NewRequest: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		response, err := s.httpClient.Do(request)
		if err != nil {
			log.Printf("httpClient.Do: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}
		if response.StatusCode != http.StatusOK {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}
		_ = response.Body.Close()

		err = s.client.Publish("", taskAssignedEventType, taskAssigned)
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

		taskCost, err := s.storage.GetTaskCostByTaskID(taskID)
		if err != nil {
			log.Printf("storage.GetTaskCostByTaskID: %s\n", err.Error())
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Create exchange message in a queue
		taskCompleted := TaskCompletedOut{
			Amount:     taskCost.CompleteCost,
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

			var taskCost *TaskCost
			taskCost, err = s.storage.GetTaskCostByTaskID(task.ID)
			if err != nil {
				log.Printf("storage.GetTaskCostByTaskID: %s\n", err.Error())
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			// Create exchange message in a queue
			taskAssigned := TaskAssignedOut{
				Amount:     taskCost.AssignCost,
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
