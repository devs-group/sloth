-- +goose Up
CREATE TABLE IF NOT EXISTS organisation_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NOT NULL,
    invitation_token VARCHAR(1024) NOT NULL,
    valid_until DATETIME NOT NULL,

    -- Foreign Keys
    organisation_id INTEGER NOT NULL,

    CONSTRAINT FK_Invitation_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,

    CONSTRAINT UQ_Email_Organisation UNIQUE(email, organisation_id),

    CONSTRAINT CK_EmailNotEmpty CHECK (email <> '')
);

-- +goose Down
DROP TABLE IF EXISTS organisation_invitations;