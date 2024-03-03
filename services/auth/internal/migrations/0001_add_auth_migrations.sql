-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    username   VARCHAR(50) NOT NULL,
    password   VARCHAR(50) NOT NULL,
    role       VARCHAR(10) NOT NULL
);

-- +goose Down
DROP TABLE users;
