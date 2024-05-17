-- +goose Up
CREATE TABLE IF NOT EXISTS organisations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id VARCHAR(255),
    name VARCHAR(255),
    UNIQUE(owner_id, name)
);

CREATE TABLE IF NOT EXISTS organisation_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organisation_id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    invitation_token VARCHAR(1024) NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(email, organisation_id),
    CONSTRAINT fk_organisation_invitations
        FOREIGN KEY (organisation_id)
        REFERENCES organisations(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organisation_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organisation_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    UNIQUE(organisation_id, user_id),
    CONSTRAINT fk_organisation_members
        FOREIGN KEY (organisation_id)
        REFERENCES organisations(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS projects_in_organisations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    organisation_id INTEGER NOT NULL,
    CONSTRAINT fk_projects_organisations
        FOREIGN KEY (organisation_id)
        REFERENCES organisations(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_projects_organisations
        FOREIGN KEY (project_id) 
        REFERENCES projects(id) 
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS projects_in_organisations;
DROP TABLE IF EXISTS organisation_members;
DROP TABLE IF EXISTS organisation_invitations;
DROP TABLE IF EXISTS organisations;