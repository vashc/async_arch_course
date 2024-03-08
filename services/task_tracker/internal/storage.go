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
INSERT INTO users(id, username, role)
VALUES (?, ?, ?);
`

	tx, err := s.sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.InsertBySql(
		query,
		user.ID,
		user.Username,
		user.Role,
	).Load(user)
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (s *Storage) CreateTask(task *Task) error {
	query := `
INSERT INTO tasks(description, status, author_id, assignee_id)
VALUES (?, ?, ?, ?)
RETURNING id;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.InsertBySql(
		query,
		task.Description,
		task.Status,
		task.AuthorID,
		task.AssigneeID,
	).Load(task)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) UpdateTaskStatus(taskID uuid.UUID, status TaskStatus) error {
	query := `
UPDATE tasks
SET status = ?
WHERE id = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.UpdateBySql(
		query,
		status,
		taskID,
	).Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) UpdateTaskAssignee(taskID, assigneeID uuid.UUID) error {
	query := `
UPDATE tasks
SET assignee_id = ?
WHERE id = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.UpdateBySql(
		query,
		assigneeID,
		taskID,
	).Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetTaskByID(id uuid.UUID) (task *Task, err error) {
	query := `
SELECT *
FROM tasks
WHERE id = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	err = tx.SelectBySql(query, id).LoadOne(&task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Storage) GetUsersByRole(role Role) (users []*User, err error) {
	query := `
SELECT *
FROM users
WHERE role = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	users = make([]*User, 0)

	_, err = tx.SelectBySql(query, role).Load(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Storage) GetTasksByStatus(status TaskStatus) (tasks []*Task, err error) {
	query := `
SELECT *
FROM tasks
WHERE status = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	tasks = make([]*Task, 0)

	_, err = tx.SelectBySql(query, status).Load(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) GetTasksByAssigneeID(assigneeID uuid.UUID) (tasks []*Task, err error) {
	query := `
SELECT *
FROM tasks
WHERE assignee_id = ?;
`

	tx, err := s.sess.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	tasks = make([]*Task, 0)

	_, err = tx.SelectBySql(query, assigneeID).Load(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
