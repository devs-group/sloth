package services

import (
	"database/sql"
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

	// populated internal
	GothUser *goth.User `json:"-"`
}

type AuthMethod struct {
	AuthID       int     `json:"auth_id" db:"auth_id"` // Primary Key
	UserID       int     `json:"user_id" db:"user_id"` // Foreign Key
	MethodType   string  `json:"method_type" db:"method_type"`
	PasswordHash *string `json:"password_hash,omitempty" db:"password_hash"`
	SocialID     *string `json:"social_id,omitempty" db:"social_id"`
}

func (g *User) GetUserWithSocialID(socialID string, tx *sqlx.Tx) (bool, error) {
	query := "SELECT * FROM auth_methods a LEFT JOIN users u ON a.user_id = u.user_id WHERE a.social_id=$1"
	var user User
	if err := tx.Get(&user, query, socialID); err != nil {
		return false, err
	}

	*g = user

	return true, nil
}

func GetUserByMail(email string, tx *sqlx.Tx) (*User, error) {
	query := "SELECT u.* FROM users u JOIN auth_methods a ON a.user_id = u.user_id WHERE u.email=$1 LIMIT 1"
	var user User
	if err := tx.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func UpsertUserBySocialIDAndMethod(methodType string, user *goth.User, tx *sqlx.Tx) (int, error) {
	var userID int
	query := `SELECT user_id FROM auth_methods WHERE social_id=$1 AND method_type=$2`
	err := tx.Get(&userID, query, user.UserID, methodType)
	if err == nil {
		// This means we already have this user with this social login
		return userID, nil
	}

	var email string
	emailIsVerified := false
	if user.Email != "" {
		emailIsVerified = true
		email = user.Email
	}

	// The user can still exists, just not with the given methodType
	existingUser, err := GetUserByMail(email, tx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, errors.Wrap(err, "can't select existing user")
		}
	}

	// Only create the user if we can't find an entry
	if existingUser == nil {
		query = `INSERT INTO users (email, username, email_verified) VALUES( $1, $2, $3 ) RETURNING user_id;`
		err = tx.Get(&userID, query, email, user.NickName, emailIsVerified)
		if err != nil {
			return 0, errors.Wrap(err, "can't insert new user")
		}
	} else {
		// Otherwise make sure we set the existing user ID
		userID = existingUser.UserID
	}

	// We always insert the auth_method which happens only once per user and method
	query = `INSERT INTO auth_methods( user_id, method_type, social_id ) VALUES ( $1, $2, $3 );`
	if _, err = tx.Exec(query, userID, methodType, user.UserID); err != nil {
		return 0, errors.Wrap(err, "can't insert auth method")
	}

	return userID, nil
}
