-- +goose Up
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NULL,
    username VARCHAR(255) NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE( email )
);

CREATE TABLE auth_methods (
    auth_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    method_type VARCHAR(50) NOT NULL,  -- e.g., 'github', 'google', 'email_password'
    social_id VARCHAR(255) NULL,  -- NULL for email/password
    password_hash VARCHAR(255) NULL,  -- NULL for social logins
    UNIQUE(method_type,social_id),  -- There is always only 1 combination of social login and id 
    UNIQUE(user_id, method_type),   -- There exists no duplicate method types for a single user 
                                    -- user1: ( google, google ) wrong
                                    -- user1: ( google, github ) correct
    CONSTRAINT fk_auth_methods 
        FOREIGN KEY (user_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_auth_methods_user_id ON auth_methods (user_id);
CREATE INDEX idx_auth_methods_method_type ON auth_methods (method_type);
CREATE INDEX idx_auth_methods_social_id ON auth_methods (social_id);

-- +goose Down
DROP TABLE auth_methods;
DROP TABLE users;
