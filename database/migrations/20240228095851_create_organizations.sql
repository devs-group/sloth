-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   user_id VARCHAR(255),
   name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS organization_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    CONSTRAINT fk_organizations_invations 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    CONSTRAINT fk_organization_members 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS temp_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    project_id INTEGER NOT NULL,
    dcj JSON
);
INSERT INTO temp_services ( id, name, project_id, dcj )
    SELECT * FROM services;


CREATE TABLE IF NOT EXISTS temp_projects( 
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unique_name VARCHAR(255),
    access_token VARCHAR(255),
    name VARCHAR(255),
    user_id VARCHAR(255),
    path VARCHAR(255),
    organization_id INTEGER DEFAULT NULL
);

INSERT INTO temp_projects ( id, unique_name, access_token, name, user_id, path )
    SELECT * FROM projects;

DROP TABLE IF EXISTS projects;

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unique_name VARCHAR(255),
    access_token VARCHAR(255),
    name VARCHAR(255),
    user_id VARCHAR(255),
    path VARCHAR(255),
    organization_id INTEGER DEFAULT NULL,
    CONSTRAINT fk_project_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE RESTRICT
);

INSERT INTO projects ( id, unique_name, access_token, name, user_id, path, organization_id )
    SELECT * FROM temp_projects;

INSERT INTO services ( id, name, project_id, dcj )
    SELECT * FROM temp_services;

DROP TABLE IF EXISTS temp_projects;
DROP TABLE IF EXISTS temp_services;

-- +goose Down
DROP CONSTRAINT fk_organizations_projects;
ALTER TABLE projects DROP COLUMN organization_id;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organization_invitations;
DROP TABLE IF EXISTS organizations;