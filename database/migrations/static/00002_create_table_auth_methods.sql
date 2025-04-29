-- +goose Up
CREATE TABLE auth_methods (
    auth_id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- e.g. 'github', 'google', 'email_password'
    method_type VARCHAR(50) NOT NULL,
    -- NULL when email/password
    social_id VARCHAR(255) NULL,
    -- NULL when social logins
    password_hash VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign Keys
    user_id INTEGER NOT NULL,

    CONSTRAINT FK_AuthMethod_User FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,

    CONSTRAINT UQ_MethodType_SocialId UNIQUE(method_type, social_id),
    CONSTRAINT UQ_MethodType_User UNIQUE(method_type, user_id)
);

CREATE INDEX IDX_AuthMethod_UserID ON auth_methods (user_id);
CREATE INDEX IDX_AuthMethod_MethodType ON auth_methods (method_type);
CREATE INDEX IDX_AuthMethod_SocialID ON auth_methods (social_id);

-- +goose Down
DROP TABLE IF EXISTS auth_methods;