-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    path VARCHAR(255),
    unique_name VARCHAR(255),
    access_token VARCHAR(255),

    -- Foreign Keys
    organisation_id INTEGER NOT NULL,

    CONSTRAINT FK_Project_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS projects;
