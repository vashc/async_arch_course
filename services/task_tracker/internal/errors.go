package internal

import "errors"

var (
	errDbrOpenConnection = errors.New("dbr failed to create connection")

	errGooseSetDialect   = errors.New("goose failed to set up dialect")
	errGooseUpMigrations = errors.New("goose failed to up migrations")
)

var (
	ErrUnsupportedMediaType = errors.New("Content-Type header is not application/json")
	ErrRequestBodyDeconding = errors.New("request body contains badly formed JSON")
	ErrUnathorizedUser      = errors.New("unauthorized user")
	ErrWrongSignMethod      = errors.New("incorrect sign method")
)
