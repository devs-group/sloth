-- +goose Up
CREATE TABLE IF NOT EXISTS docker_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255),
    password VARCHAR(255),
    registry VARCHAR(255),

    -- Foreign Keys
    project_id INTEGER NOT NULL,

    CONSTRAINT FK_DockerCredential_Project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS docker_credentials;
