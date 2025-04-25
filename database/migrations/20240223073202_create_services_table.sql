-- +goose Up
CREATE TABLE IF NOT EXISTS services (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(255) NOT NULL,
  project_id INTEGER NOT NULL,
  dcj JSON,
  CONSTRAINT fk_services_project_id FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS services;
