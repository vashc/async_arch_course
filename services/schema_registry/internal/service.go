package internal

import (
	"context"
	"io"
	"time"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/xeipuuv/gojsonschema"
)

func NewService(config *Config) *Service {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", config.API.Host, config.API.Port),
		ReadHeaderTimeout: time.Second * 5,
	}

	service := &Service{
		config: config,
		server: server,
		Mux:    chi.NewRouter(),
	}

	service.server.Handler = service

	return service
}

func (s *Service) InstantiateRoutes() {
	s.Use(
		middleware.Timeout(5*time.Second),
		LogRequest,
	)

	s.Route("/validate", func(router chi.Router) {
		router.Post(fmt.Sprintf("/{%s}/event", requestParamEventType), s.validateEventHandler())
	})

	s.Get("/health", s.healthHandler())
}

func (s *Service) Start() error {
	return s.server.ListenAndServe()
}

func (s *Service) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) validateEventHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		requestEventType := eventType(chi.URLParam(r, requestParamEventType))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		if s.isValidEvent(requestEventType, body) {
			w.WriteHeader(http.StatusOK)

			resp, err := json.Marshal(Response{Status: http.StatusText(http.StatusOK)})
			if err != nil {
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			_, _ = w.Write(resp)
			return
		}

		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
	}
}

func (s *Service) isValidEvent(requestEventType eventType, body []byte) bool {
	schemaLoader := gojsonschema.NewReferenceLoader(
		fmt.Sprintf("../schemas/{%s}/1.json", requestEventType),
	)
	bodyLoader := gojsonschema.NewBytesLoader(body)

	result, err := gojsonschema.Validate(schemaLoader, bodyLoader)
	if err != nil {
		return false
	}

	if result.Valid() {
		return true
	}

	return false
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
