-- +goose Up
CREATE TABLE IF NOT EXISTS organisations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT CK_NameNotEmpty CHECK (name <> '')
);

-- +goose Down
DROP TABLE IF EXISTS organisations;