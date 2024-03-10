-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tasks_costs (
    id              UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    task_id         UUID        NOT NULL,
    assign_cost     INT         NOT NULL,
    complete_cost   INT         NOT NULL,

    CONSTRAINT fk_tasks_costs_to_tasks FOREIGN KEY (task_id) REFERENCES tasks(id) ON UPDATE RESTRICT
);

-- +goose Down
DROP TABLE tasks_costs;
