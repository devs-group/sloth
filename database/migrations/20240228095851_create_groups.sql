-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id VARCHAR(255),
    name VARCHAR(255),
    UNIQUE(owner_id, name)
);

CREATE TABLE IF NOT EXISTS group_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    CONSTRAINT fk_groups_invations 
        FOREIGN KEY (group_id) 
        REFERENCES groups(id) 
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    UNIQUE(group_id, user_id),
    CONSTRAINT fk_group_members 
        FOREIGN KEY (group_id) 
        REFERENCES groups(id) 
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
    group_id INTEGER DEFAULT NULL
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
    group_id INTEGER DEFAULT NULL,
    CONSTRAINT fk_project_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE RESTRICT
);

INSERT INTO projects ( id, unique_name, access_token, name, user_id, path, group_id )
    SELECT * FROM temp_projects;

INSERT INTO services ( id, name, project_id, dcj )
    SELECT * FROM temp_services;

DROP TABLE IF EXISTS temp_projects;
DROP TABLE IF EXISTS temp_services;

-- +goose Down
DROP CONSTRAINT fk_groups_projects;
ALTER TABLE projects DROP COLUMN group_id;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS group_invitations;
DROP TABLE IF EXISTS groups;