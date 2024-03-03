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
	"github.com/go-chi/jwtauth/v5"
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
	)

	s.Route("/user", func(router chi.Router) {
		router.Get(
			fmt.Sprintf("/{%s}", requestParamUserID),
			s.getUserHandler(),
		)
		router.Post("/create", s.createUserHandler())
		router.Post("/auth", s.authUserHandler())
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

func (s *Service) createUserHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := new(User)

		err := BodyParser(w, r, user)
		if err != nil {
			code := http.StatusUnprocessableEntity
			http.Error(w, http.StatusText(code), code)
			return
		}

		if err = s.storage.CreateUser(user); err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		resp, err := json.Marshal(CreateUserResponse{ID: user.ID})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Create exchange message in a queue
		userCreated := UserCreatedOut{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		}

		err = s.client.Publish("", userCreated)
		if err != nil {
			log.Printf("client.Publish: %s\n", err.Error())
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) authUserHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(AuthRequest)

		err := BodyParser(w, r, req)
		if err != nil {
			code := http.StatusUnprocessableEntity
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Check if we have such a user in our DB
		var user *User
		user, err = s.storage.GetUserByUsername(req.Username, req.Password)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, dbr.ErrNotFound) {
				code = http.StatusNotFound
			}

			http.Error(w, http.StatusText(code), code)
			return
		}

		token := jwtauth.New(authTokenAlgo, []byte(s.config.JWTSecret), nil)

		var tokenString string
		_, tokenString, err = token.Encode(AuthToken{requestParamUserID: user.ID})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		w.Header().Set(HeaderAuth, fmt.Sprintf("%s%s", HeaderBearer, tokenString))

		resp, err := json.Marshal(Response{Status: http.StatusText(http.StatusOK)})
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = w.Write(resp)
	}
}

func (s *Service) getUserHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		userID, err := uuid.Parse(chi.URLParam(r, requestParamUserID))
		if err != nil {
			log.Printf("uuid.Parse: %s\n", err.Error())
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		var user *User
		user, err = s.storage.GetUserByID(userID)
		if err != nil {
			log.Printf("storage.GetUserByID: %s\n", err.Error())
			code := http.StatusInternalServerError
			if errors.Is(err, dbr.ErrNotFound) {
				code = http.StatusNotFound
			}
			http.Error(w, http.StatusText(code), code)
			return
		}

		var resp []byte
		resp, err = json.Marshal(user)
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
