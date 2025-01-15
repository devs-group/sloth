-- +goose Up
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) DEFAULT '',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT UQ_Email UNIQUE(email),

    CONSTRAINT CK_EmailNotEmpty CHECK (email <> '')
);

CREATE INDEX IDX_User_Email ON users (email);

-- +goose Down
DROP TABLE IF EXISTS users;
