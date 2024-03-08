package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gocraft/dbr/v2"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type Service struct {
	config  *Config
	server  *http.Server
	storage *Storage
	client  *RabbitClient

	*chi.Mux
}

type Storage struct {
	sess *dbr.Session
}

type DB struct {
	Host string `envconfig:"DB_HOST" required:"true" default:"postgres_auth"`
	Port string `envconfig:"DB_PORT" required:"true" default:"5432"`
	Name string `envconfig:"DB_NAME" required:"true" default:"main"`
	User string `envconfig:"DB_USER" required:"true" default:"user"`
	Pass string `envconfig:"DB_PASS" required:"true" default:"pass"`
}

func (db *DB) uri() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		db.User,
		db.Pass,
		db.Host,
		db.Port,
		db.Name,
	)
}

type RabbitConfig struct {
	RabbitHost  string `envconfig:"RABBIT_HOST" required:"true" default:"event_bus"`
	RabbitPort  string `envconfig:"RABBIT_PORT" required:"true" default:"5672"`
	RabbitLogin string `envconfig:"RABBIT_LOGIN" required:"true" default:"user"`
	RabbitPass  string `envconfig:"RABBIT_PASS" required:"true" default:"pass"`
}

func (rc *RabbitConfig) uri() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/",
		RabbitProtocol,
		rc.RabbitLogin,
		rc.RabbitPass,
		rc.RabbitHost,
		rc.RabbitPort,
	)
}

type API struct {
	Host string `envconfig:"AUTH_HOST" required:"true" default:"0.0.0.0"`
	Port string `envconfig:"AUTH_PORT" required:"true" default:"8000"`
}

type Config struct {
	DB       DB
	EventBus RabbitConfig
	API      API

	JWTSecret string `envconfig:"JWT_SECRET" required:"true" default:"some_default_jwt_secret"`
}

type CreateUserResponse struct {
	ID uuid.UUID `json:"id"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Status string `json:"status"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      Role      `json:"role"`
}

type RabbitClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

type AuthToken map[string]interface{}
