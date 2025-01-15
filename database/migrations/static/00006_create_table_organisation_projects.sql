-- +goose Up
CREATE TABLE IF NOT EXISTS projects_2_organisations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Foreign Keys
    project_id INTEGER NOT NULL,
    organisation_id INTEGER NOT NULL,

    CONSTRAINT FK_Project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT FK_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS projects_2_organisations;