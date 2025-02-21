-- +goose Up
CREATE TABLE IF NOT EXISTS organisation_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role VARCHAR(32) NOT NULL,

    -- Foreign Keys
    organisation_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,

    CONSTRAINT FK_OrganisationMember_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    CONSTRAINT FK_OrganisationMember_User FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,

    CONSTRAINT UQ_Organisation_User UNIQUE(organisation_id, user_id),

    CONSTRAINT CK_RoleValid CHECK (role IN ('owner', 'admin', 'member'))
);

-- +goose Down
DROP TABLE IF EXISTS organisation_members;