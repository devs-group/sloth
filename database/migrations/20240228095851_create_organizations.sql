-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id VARCHAR(255),
    name VARCHAR(255),
    UNIQUE(owner_id, name)
);

CREATE TABLE IF NOT EXISTS organization_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    invitation_token VARCHAR(1024) NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(email, organization_id),
    CONSTRAINT fk_organization_invitations 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    UNIQUE(organization_id, user_id),
    CONSTRAINT fk_organization_members 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS projects_in_organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    CONSTRAINT fk_projects_organizations
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_projects_organizations
        FOREIGN KEY (project_id) 
        REFERENCES projects(id) 
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS projects_in_organizations;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organization_invitations;
DROP TABLE IF EXISTS organizations;