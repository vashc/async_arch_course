package internal

import (
	"embed"
	"fmt"
	"log"

	"github.com/gocraft/dbr/v2"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // Driver
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var embedFS embed.FS

func NewStorage(config *Config) (*Storage, error) {
	conn, err := newConn(config)
	if err != nil {
		return nil, err
	}

	sess := conn.NewSession(conn.EventReceiver)

	if err := runMigrations(sess); err != nil {
		return nil, err
	}

	return &Storage{sess: sess}, nil
}

func newConn(cfg *Config) (*dbr.Connection, error) {
	log.Printf("Establishing connection to %s", cfg.DB.uri())
	conn, err := dbr.Open(dbDriver, cfg.DB.uri(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errDbrOpenConnection, err.Error())
	}

	return conn, nil
}

func runMigrations(session *dbr.Session) error {
	goose.SetBaseFS(embedFS)

	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("%w: %s", errGooseSetDialect, err.Error())
	}

	err = goose.Up(session.DB, "migrations")
	if err != nil {
		return fmt.Errorf("%w: %s", errGooseUpMigrations, err.Error())
	}

	return nil
}

func (s *Storage) Close() error {
	return s.sess.Close()
}

func (s *Storage) CreateUser(user *User) error {
	query := `
INSERT INTO users(username, password, role)
VALUES (?, ?, ?)
RETURNING id;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.InsertBySql(
		query,
		user.Username,
		user.Password,
		user.Role,
	).Load(user)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetUserByUsername(username, password string) (user *User, err error) {
	query := `
SELECT *
FROM users
WHERE username = ? AND password = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.SelectBySql(query, username, password).LoadOne(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Storage) GetUserByID(id uuid.UUID) (user *User, err error) {
	query := `
SELECT *
FROM users
WHERE id = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.SelectBySql(query, id).LoadOne(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
