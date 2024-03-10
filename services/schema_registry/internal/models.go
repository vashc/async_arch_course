package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Service struct {
	config *Config
	server *http.Server

	*chi.Mux
}

type API struct {
	Host string `envconfig:"SCHEMA_REGISTRY_HOST" required:"true" default:"0.0.0.0"`
	Port string `envconfig:"SCHEMA_REGISTRY_PORT" required:"true" default:"8010"`
}

type Config struct {
	API API

	JWTSecret string `envconfig:"JWT_SECRET" required:"true" default:"some_default_jwt_secret"`
}

type Response struct {
	Status string `json:"status"`
}
