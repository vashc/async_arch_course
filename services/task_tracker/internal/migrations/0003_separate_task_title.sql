-- +goose Up

ALTER TABLE tasks ADD COLUMN jira_id TEXT NOT NULL DEFAULT '';

ALTER TABLE tasks ADD COLUMN title TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tasks DROP COLUMN jira_id;

ALTER TABLE tasks DROP COLUMN title;
