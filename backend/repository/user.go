package repository

import (
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
	"github.com/pkg/errors"
)

type User struct {
	UserID        int       `json:"user_id" db:"user_id"`
	Email         *string   `json:"email,omitempty" db:"email"`
	UserName      *string   `json:"username" db:"username"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type AuthMethod struct {
	AuthID       int     `json:"auth_id" db:"auth_id"` // Primary Key
	UserID       int     `json:"user_id" db:"user_id"` // Foreign Key
	MethodType   string  `json:"method_type" db:"method_type"`
	PasswordHash *string `json:"password_hash,omitempty" db:"password_hash"`
	SocialID     *string `json:"social_id,omitempty" db:"social_id"`
}

func (g *User) GetUserWithSocialID(socialID string, tx *sqlx.Tx) (bool, error) {
	query := "SELECT * FROM auth_methods a LEFT JOIN user u ON a.user_id = u.user_id WHERE a.social_id=$1"
	var user User
	if err := tx.Get(&user, query, socialID); err != nil {
		return false, err
	}

	*g = user

	return true, nil
}

func UpsertUserByUserID(user *goth.User, tx *sqlx.Tx) (bool, error) {
	var userID int
	query := `INSERT OR IGNORE INTO users (email, username, email_verified)
    VALUES ($1, $2, $3) RETURNING user_id;
    `

	err := tx.Get(&userID, query, user.Email, user.NickName, false)
	if userID == 0 {
		// User already exists
		query = "SELECT user_id FROM users WHERE email=$1"
		err = tx.Get(&userID, query, user.Email)
	}

	slog.Info("USERID", "USER", user)

	if err != nil {
		return false, errors.Wrap(err, "can't insert new user")
	}

	query = `INSERT OR IGNORE INTO auth_methods( 
      user_id, 
      method_type, 
      social_id
    )
  VALUES
    (
      $1,
      "github", 
      $2
    );`

	slog.Info("TEST", "TEST2", userID, "socialID", user.UserID)
	if _, err = tx.Exec(query, userID, user.UserID); err != nil {
		slog.Error("ERROR", "found ERROR", err)
		return false, err
	}
	slog.Info("DONE", "USER STORED", "STORED")
	return true, nil
}
