package internal

import (
	"context"
	"log"
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

	s.Route("/", func(router chi.Router) {
		router.Get("/operation/log", s.getOperationLogHandler())
		router.Get("/balance", s.getBalanceHandler())
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

//nolint:dupl // It has different logic
func (s *Service) getOperationLogHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(requestParamUserID).(uuid.UUID)

		user, err := s.storage.GetUserByID(userID)
		if err != nil {
			log.Printf("storage.GetUserByID: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		if user.Role == accountantRole || user.Role == adminRole {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			return
		}

		operations, err := s.storage.GetOperationsByUserID(userID)
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

		resp, err := json.Marshal(operations)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

//nolint:dupl // It has different logic
func (s *Service) getBalanceHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(requestParamUserID).(uuid.UUID)

		user, err := s.storage.GetUserByID(userID)
		if err != nil {
			log.Printf("storage.GetUserByID: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		if user.Role == accountantRole || user.Role == adminRole {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			return
		}

		account, err := s.storage.GetAccountByUserID(userID)
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

		resp, err := json.Marshal(account)
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
