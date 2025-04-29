-- +goose Up
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) DEFAULT '',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    current_organisation_id INTEGER NOT NULL,

    CONSTRAINT FK_User_Organisation FOREIGN KEY (current_organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,

    CONSTRAINT UQ_Email UNIQUE(email),

    CONSTRAINT CK_EmailNotEmpty CHECK (email <> '')
);

CREATE INDEX IDX_User_Email ON users (email);

-- +goose Down
DROP TABLE IF EXISTS users;
