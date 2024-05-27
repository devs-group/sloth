package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
)

type Organisation struct {
	ID      int      `json:"id" db:"id"`
	Name    string   `json:"organisation_name" db:"name" binding:"required"`
	OwnerID string   `json:"-" db:"owner_id"`
	IsOwner bool     `json:"is_owner" db:"is_owner"`
	Members []string `json:"members"`
}

type Invitation struct {
	Email            string `json:"email" db:"email" binding:"required"`
	OrganisationName string `json:"organisation_name" db:"name" binding:"required"`
}

type AcceptInvite struct {
	OrganisationID int       `json:"organisation_id" db:"organisation_id"`
	TimeStamp      time.Time `json:"timestamp" db:"timestamp"`
}

type OrganisationProjects struct {
	UniqueName  string `json:"upn" db:"unique_name"`
	ProjectName string `json:"name" db:"name"`
	ID          int    `json:"id"   db:"id"`
}

func (o *Organisation) CreateOrganisation(tx *sqlx.Tx) error {
	var oID int
	query := `INSERT INTO organisations( name, owner_id ) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&oID, query, o.Name, o.OwnerID); err != nil {
		return err
	}

	var mID int
	query = `INSERT INTO organisation_members (organisation_id, user_id) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&mID, query, oID, o.OwnerID); err != nil {
		return err
	}

	return nil
}

// SelectOrganisations returns a list of the user's Organisations
// User must be the owner of the organisation and also be a member of it.
func SelectOrganisations(userID string, tx *sqlx.Tx) ([]Organisation, error) {
	organisations := make([]Organisation, 0)
	query := `SELECT o.id,
					 o.name, 
					 o.owner_id = om.user_id as is_owner
				FROM organisations o
				JOIN organisation_members om ON o.id = om.organisation_id
				WHERE om.user_id = $1;
	`
	err := tx.Select(&organisations, query, userID)
	if err != nil {
		return nil, err
	}
	return organisations, nil
}

// SelectOrganisation updates the Organisation struct with its members. Only if the User is the Owner of the Organisation
// The Organisation struct (o) must have the 'Name' and 'OwnerID' fields set before calling this function,
// as these fields are used to identify the specific organisation:
//
//   - o.Name: The name of the organisation for which member user IDs are to be retrieved.
//   - o.OwnerID: The ID of the owner of the organisation. Only the Owner can retrieve a list of members of the organisation
func SelectOrganisation(tx *sqlx.Tx, orgID int, userID string) (*Organisation, error) {
	var result int
	query := `SELECT 1
    FROM organisation_members om
    WHERE om.organisation_id = $1 AND om.user_id = $2`
	row := tx.QueryRow(query, orgID, userID)
	err := row.Scan(&result)
	if err != nil {
		return nil, err
	}
	if result == 0 {
		return nil, errors.New("organisation not found")
	}

	query = `SELECT o.id, 
    				o.name,
    				o.owner_id,
    				om.user_id
				FROM organisations o
				INNER JOIN organisation_members om ON om.organisation_id = o.id
				WHERE o.id = $1;
	`
	rows, err := tx.Query(query, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organisation Organisation
	memberIDs := make([]string, 0)
	isOwner := false
	for rows.Next() {
		var ownerID string
		var memberID string
		err = rows.Scan(&organisation.ID, &organisation.Name, &ownerID, &memberID)
		memberIDs = append(memberIDs, memberID)
		if ownerID == userID {
			isOwner = true
		}
	}
	if organisation.ID == 0 {
		return nil, errors.New("organisation not found")
	}
	organisation.Members = memberIDs
	organisation.IsOwner = isOwner

	return &organisation, nil
}

func (o *Organisation) DeleteOrganisation(tx *sqlx.Tx) error {
	// TODO: Remove project's from organisations
	query := `DELETE FROM organisations WHERE owner_id = $1 AND id = $2`
	res, err := tx.Exec(query, o.OwnerID, o.ID)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("found different amounts of organisations for user %s", o.OwnerID)
	}
	return nil
}

func DeleteMember(ownerID, memberID, organisationName string, tx *sqlx.Tx) error {
	query := `
		DELETE FROM organisation_members
		WHERE user_id = $1 AND organisation_id = (
		SELECT id FROM organisations
		WHERE owner_id = $2 AND name = $3
		) AND user_id <> $2;
	`
	res, err := tx.Exec(query, memberID, ownerID, organisationName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 member from organisations '%s', but deleted %d", organisationName, rem)
	}

	return nil
}

func PutInvitation(ownerID, newMemberID, organisationName, invitationToken string, tx *sqlx.Tx) error {
	if isAlreadyMember := CheckIsMemberOfOrganisation(newMemberID, organisationName, tx); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of organisation: %s", newMemberID, organisationName)
	}

	query := `
	INSERT INTO organisation_invitations(organisation_id, email, invitation_token )
		SELECT id, $1, $4 FROM organisations WHERE owner_id = $2 AND name = $3;
	`
	res, err := tx.Exec(query, newMemberID, ownerID, organisationName, invitationToken)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organisations '%s', but added %d", organisationName, rem)
	}

	return nil
}

func PutMember(newMemberID, organisationName string, tx *sqlx.Tx) error {
	query := `
    INSERT INTO organisation_members(organisation_id, user_id)
    	SELECT organisation_id, $1 FROM organisation_invitations oi
		JOIN organisations o ON o.id = oi.organisation_id WHERE oi.user_id = $2 AND o.name = $3;
    `
	res, err := tx.Exec(query, newMemberID, organisationName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organisations '%s', but added %d", organisationName, rem)
	}

	return nil
}

func GetInvitations(userID string, tx *sqlx.Tx) ([]Invitation, error) {
	invites := make([]Invitation, 0)
	query := `SELECT oi.email, o.name 
				FROM organisation_invitations oi 
				JOIN organisations o ON o.id = oi.organisation_id
				WHERE oi.id = $1
				ORDER BY oi.id DESC;
	`
	if err := tx.Select(&invites, query, userID); err != nil {
		return nil, err
	}
	return invites, nil
}

func CheckIsMemberOfOrganisation(userID, organisationName string, tx *sqlx.Tx) bool {
	query := `SELECT 1 FROM organisation_members om 
			  JOIN organisations o ON o.id = om.organisation_id  WHERE user_id = $1 AND o.name = $2;`
	isMemberOfSomeOrganisation := false
	_ = tx.Get(&isMemberOfSomeOrganisation, query, userID, organisationName)
	return isMemberOfSomeOrganisation
}

func GetInvitation(email, token string, tx *sqlx.Tx) (*Invitation, error) {
	query := `SELECT oi.email, o.name 
	FROM organisation_invitations oi 
	JOIN organisations o ON o.id = oi.organisation_id 
	WHERE oi.email=$1 AND oi.invitation_token=$2`

	var invitation Invitation
	err := tx.Get(&invitation, query, email, token)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func AcceptInvitation(userID, email, token string, tx *sqlx.Tx) (bool, error) {
	var accept AcceptInvite

	query := `SELECT timestamp, organisation_id FROM organisation_invitations WHERE email = $1 AND invitation_token = $2`
	err := tx.Get(&accept, query, email, token)
	if err != nil {
		slog.Info("Error", "can't find invitation", err)
		return false, err
	}

	query = `DELETE FROM organisation_invitations WHERE email=$1 AND invitation_token=$2 RETURNING organisation_id`
	err = tx.Get(&accept, query, email, token)
	if err != nil {
		slog.Info("Error", "cant delete entry ", err)
		return false, err
	}

	if time.Since(accept.TimeStamp) > config.EmailInvitationMaxValid {
		slog.Info("Error", "The invitation is too old", accept)
		return false, fmt.Errorf("can't accept invitation, invitation too old")
	}

	query = `INSERT INTO organisation_members ( organisation_id, user_id ) VALUES ( $1, $2 )`
	res, err := tx.Exec(query, accept.OrganisationID, userID)
	if err != nil {
		slog.Info("Error", "can't insert new member to organisation", err)
		return false, err
	}

	ins, err := res.RowsAffected()
	if err != nil {
		slog.Info("Error", "cant fetch inserted rows", err)
		return false, err
	}

	if ins != 1 {
		slog.Info("Error", "inserted multiple users to organisation", "restricted")
		return false, fmt.Errorf("inserted multiple users to organisation")
	}

	return true, nil
}

func GetProjectsByOrganisationID(userID string, organisationID int, tx *sqlx.Tx) ([]OrganisationProjects, error) {
	projects := make([]OrganisationProjects, 0)
	query := `SELECT DISTINCT p.unique_name, p.name, p.id
    FROM projects p
    JOIN projects_in_organisations pio ON p.id = pio.project_id
    JOIN organisations o ON pio.organisation_id = o.id
    JOIN organisation_members om ON o.id = om.organisation_id
    WHERE o.id = $1 AND om.user_id = $2;
    `

	err := tx.Select(&projects, query, organisationID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return projects, nil
		}
		return nil, err
	}
	return projects, nil
}

func AddOrganisationProjectByUPN(userID string, organisationID int, upn string, tx *sqlx.Tx) (bool, error) {
	relationID := 0
	query := `
	INSERT INTO projects_in_organisations (project_id, organisation_id)
		SELECT p.id, $1
		FROM projects p
		WHERE p.unique_name = $2 AND p.user_id = $3 AND EXISTS (
    		SELECT 1 FROM organisations WHERE id = $1
		)
	RETURNING id;
	`
	err := tx.Get(&relationID, query, organisationID, upn, userID)
	if err != nil {
		return false, errors.Join(err, fmt.Errorf("can't add project your not owner or this project does not exist"))
	}

	return true, nil
}
