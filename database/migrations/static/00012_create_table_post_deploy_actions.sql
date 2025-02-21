-- +goose Up
CREATE TABLE post_deploy_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shell VARCHAR(255) NOT NULL,
    command VARCHAR(255) NOT NULL,
    parameters VARCHAR(255) DEFAULT '',

    -- Foreign Key
    service_id INTEGER NOT NULL,

    CONSTRAINT FK_PostDeployAction_Service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS post_deploy_actions;