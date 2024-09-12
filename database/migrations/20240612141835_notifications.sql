-- +goose Up
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    time_stamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    subject VARCHAR(255) NOT NULL,
    content VARCHAR(255) NOT NULL,
    sender VARCHAR(255) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    notification_type VARCHAR(255) NOT NULL DEFAULT "INFO"
);

-- +goose Down
DROP TABLE IF EXISTS notifications;