-- +goose Up
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subject VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    notification_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign Keys
    user_id INTEGER NOT NULL,

    CONSTRAINT FK_Notification_User FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,

    CONSTRAINT CK_NotificationTypeValid CHECK (notification_type IN ('system', 'news'))
);

-- +goose Down
DROP TABLE IF EXISTS notifications;