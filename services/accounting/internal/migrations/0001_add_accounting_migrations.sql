-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    username   VARCHAR(50) NOT NULL,
    role       VARCHAR(10) NOT NULL
);

CREATE TABLE operations (
    id          UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    amount      INT         NOT NULL,
    user_id     UUID        NOT NULL,

    CONSTRAINT fk_operations_to_users FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE RESTRICT
);

CREATE TABLE accounts (
    id          UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    amount      INT         NOT NULL,
    user_id     UUID        NOT NULL,

    CONSTRAINT fk_accounts_to_users FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE RESTRICT
);

-- +goose Down
DROP TABLE users;
DROP TABLE operations;
