package services

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
	"github.com/pkg/errors"
)

type User struct {
	UserID                int       `json:"user_id" db:"user_id"`
	Email                 *string   `json:"email,omitempty" db:"email"`
	UserName              *string   `json:"username" db:"username"`
	EmailVerified         bool      `json:"email_verified" db:"email_verified"`
	CurrentOrganisationID int       `json:"currentOrganisationID" db:"current_organisation_id"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`

	// populated internal
	GothUser *goth.User `json:"-"`
}

type SessionIDs struct {
	UserID                int `db:"user_id"`
	CurrentOrganisationID int `db:"current_organisation_id"`
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

func UpsertUserBySocialIDAndMethod(methodType string, user *goth.User, tx *sqlx.Tx) (*SessionIDs, error) {
	var sessionIDs SessionIDs
	query := `
		SELECT am.user_id, u.current_organisation_id
		FROM auth_methods am
		JOIN users u ON u.user_id = am.user_id
		WHERE
		    am.social_id=$1
			AND am.method_type=$2
	`
	err := tx.Get(&sessionIDs, query, user.UserID, methodType)
	if err == nil {
		// This means we already have this user with this social login
		return &sessionIDs, nil
	}

	var email string
	emailIsVerified := false
	if user.Email != "" {
		emailIsVerified = true
		email = user.Email
	}

	// The user can still exist, just not with the given methodType
	existingUser, err := GetUserByMail(email, tx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "can't select existing user")
		}
	}

	// Only create the user if we can't find an entry
	if existingUser == nil {
		// Create a default organisation for the user and assign it
		var organisationID int
		query = `INSERT INTO organisations (name, is_default) VALUES( $1, $2 ) RETURNING id;`
		err = tx.Get(&organisationID, query, "My Organisation", true)
		if err != nil {
			return nil, errors.Wrap(err, "can't create organisation for user")
		}

		// Create the user
		query = `
			INSERT INTO users (email, username, email_verified, current_organisation_id)
			VALUES( $1, $2, $3, $4 )
			RETURNING user_id, current_organisation_id;
		`
		err = tx.Get(&sessionIDs, query, email, user.NickName, emailIsVerified, organisationID)
		if err != nil {
			return nil, errors.Wrap(err, "can't insert new user")
		}

		// Finally add the user as the owner to the new organisation
		query = `INSERT INTO organisation_members (organisation_id, user_id, role) VALUES ( $1, $2, $3 )`
		res, err := tx.Exec(query, sessionIDs.CurrentOrganisationID, sessionIDs.UserID, "owner")
		if err != nil {
			return nil, errors.Wrap(err, "unable to add member to organisation")
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return nil, errors.New("unable to add member to organisation")
		}
		if affected == 0 {
			return nil, errors.New("unable to add member to organisation")
		}
	} else {
		// Otherwise make sure we set the existing user ID
		sessionIDs.UserID = existingUser.UserID
		sessionIDs.CurrentOrganisationID = existingUser.CurrentOrganisationID
	}

	// We always insert the auth_method which happens only once per user and method
	query = `INSERT INTO auth_methods( user_id, method_type, social_id ) VALUES ( $1, $2, $3 );`
	if _, err = tx.Exec(query, sessionIDs.UserID, methodType, user.UserID); err != nil {
		return nil, errors.Wrap(err, "can't insert auth method")
	}

	return &sessionIDs, nil
}
