-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    username   VARCHAR(50) NOT NULL,
    role       VARCHAR(10) NOT NULL
);

CREATE TABLE tasks (
    id          UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    description TEXT        NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL,
    author_id   UUID        NOT NULL,
    assignee_id UUID        NOT NULL,

    CONSTRAINT fk_tasks_author_to_users FOREIGN KEY (author_id) REFERENCES users(id) ON UPDATE RESTRICT,
    CONSTRAINT fk_tasks_assignee_to_users FOREIGN KEY (assignee_id) REFERENCES users(id) ON UPDATE RESTRICT
);

-- +goose Down
DROP TABLE users;
DROP TABLE tasks;
