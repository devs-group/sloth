package services

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Notification struct {
	ID               int       `json:"id" db:"id"`
	Subject          string    `json:"subject" db:"subject"`
	Content          string    `json:"content" db:"content"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UserID           int       `json:"UserID" db:"user_id"`
}

func (s *S) CreateNotification(payload Notification, tx *sqlx.Tx) error {
	query := `
		INSERT INTO notifications (subject, content, notification_type, user_id) 
		VALUES ($1, $2, $3, $4);
	`
	_, err := tx.Exec(query, payload.Subject, payload.Content, payload.UserID, payload.NotificationType)
	if err != nil {
		slog.Error("Unable to store notification", "err", err)
		return err
	}

	return nil

}

func (s *S) GetNotifications(userID string, tx *sqlx.Tx) ([]Notification, error) {
	query := `
    SELECT
        n.id,
        n.subject,
        n.content,
		n.notification_type,
        n.created_at,
        n.user_id
    FROM
        notifications n
    JOIN users u ON u.user_id = $1
    WHERE
        n.user_id = u.email;
    `

	var notifications []Notification
	err := tx.Select(&notifications, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Notification{}, nil
		}
		slog.Error("Unable to get notifications", "UserID", userID, "err", err)
		return nil, err
	}

	return notifications, nil
}
