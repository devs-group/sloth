-- +goose Up
CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    usn VARCHAR(255) NOT NULL,
    dcj JSON,

    -- Foreign Keys
    project_id INTEGER NOT NULL,

    CONSTRAINT FK_Service_Project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,

    CONSTRAINT CK_NameNotEmpty CHECK (name <> ''),
    CONSTRAINT CK_USNNotEmpty CHECK (usn <> '')
);

-- +goose Down
DROP TABLE IF EXISTS services;
