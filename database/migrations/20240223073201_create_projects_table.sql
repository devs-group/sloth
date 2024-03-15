-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unique_name VARCHAR(255),
    access_token VARCHAR(255),
    name VARCHAR(255),
    user_id VARCHAR(255),
    path VARCHAR(255)
);

-- +goose Down
DROP TABLE IF EXISTS projects;
