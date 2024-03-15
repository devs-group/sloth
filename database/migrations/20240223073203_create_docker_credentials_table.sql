-- +goose Up
CREATE TABLE IF NOT EXISTS docker_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255),
    password VARCHAR(255),
    registry VARCHAR(255),
    project_id INTEGER NOT NULL,
    CONSTRAINT fk_docker_credentials_project_id
        FOREIGN KEY (project_id)
        REFERENCES projects (id)
        ON DELETE CASCADE    
);

-- +goose Down
DROP TABLE IF EXISTS docker_credentials;
