-- +goose Up
CREATE TABLE post_deploy_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL,
    parameters VARCHAR(255) NOT NULL,
    shell VARCHAR(255) NOT NULL,
    command VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS post_deploy_actions;