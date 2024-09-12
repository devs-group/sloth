package services

import (
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Notification struct {
	ID               int       `json:"id" db:"id"`
	TimeStamp        time.Time `json:"time_stamp" db:"time_stamp"`
	Recipient        string    `json:"recipient" db:"recipient"`
	Sender           string    `json:"sender" db:"sender"`
	Subject          string    `json:"subject" db:"subject"`
	Content          string    `json:"content" db:"content"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
}

func StoreNotification(userID, subject, content, recipient, notification_type string, tx *sqlx.Tx) error {
	query := `SELECT u.email FROM users u WHERE u.user_id = $1;`

	var sender string

	err := tx.Get(&sender, query, userID)
	if err != nil {
		slog.Error("Unable to find email from user")
		return err
	}

	query = `INSERT INTO notifications (subject, content, sender, recipient, notification_type) VALUES ($1, $2, $3, $4, $5);`
	_, err = tx.Exec(query, subject, content, sender, recipient, notification_type)
	if err != nil {
		slog.Error("Unable to store notification: %v", err)
		return err
	}

	return nil

}

func GetNotifications(userID string, tx *sqlx.Tx) ([]Notification, error) {
	query := `
    SELECT 
        n.id, 
        n.time_stamp, 
        n.subject, 
        n.content, 
        n.sender, 
        n.recipient,
		n.notification_type
    FROM 
        notifications n 
    JOIN 
        users u 
    ON 
        u.user_id = $1 
    WHERE 
        n.recipient = u.email;
    `

	var notifications []Notification
	err := tx.Get(&notifications, query, userID)
	if err != nil {
		slog.Error("Unable to get notifications for user: %s, err: %v", userID, err)
		return nil, err
	}

	return notifications, nil
}
