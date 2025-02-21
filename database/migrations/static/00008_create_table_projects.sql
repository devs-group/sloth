-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(255) NOT NULL,
    unique_name VARCHAR(255) NOT NULL,
    access_token VARCHAR(255) NOT NULL,

    -- Foreign Keys
    organisation_id INTEGER NOT NULL,

    CONSTRAINT FK_Project_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,

    CONSTRAINT CK_NameNotEmpty CHECK (name <> '')
);

-- +goose Down
DROP TABLE IF EXISTS projects;
