package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/models"
)

type Organisation struct {
	ID      int                  `json:"id" db:"id"`
	Name    string               `json:"organisation_name" db:"name" binding:"required"`
	OwnerID string               `json:"-" db:"owner_id"`
	IsOwner bool                 `json:"is_owner" db:"is_owner"`
	Members []OrganisationMember `json:"members"`
}

type Invitation struct {
	Email          string `json:"email" db:"email" binding:"required"`
	OrganisationID int    `json:"organisation_id" db:"organisation_id" binding:"required"`
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

type OrganisationMember struct {
	UserID   int     `json:"user_id" db:"user_id"`
	Email    *string `json:"email,omitempty" db:"email"`
	UserName *string `json:"username" db:"username"`
}

func (s *S) CreateOrganisation(o models.Organisation) (*models.Organisation, error) {
	var organisation models.Organisation
	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		query := `INSERT INTO organisations( name, owner_id ) VALUES ( $1, $2 ) RETURNING *`
		if err := tx.Get(&organisation, query, o.Name, o.OwnerID); err != nil {
			return fmt.Errorf("unable to insert organisation: %w", err)
		}

		var member models.OrganisationMember
		query = `INSERT INTO organisation_members (organisation_id, user_id) VALUES ( $1, $2 ) RETURNING *`
		if err := tx.Get(&member, query, organisation.ID, o.OwnerID); err != nil {
			return fmt.Errorf("unable to insert member: %w", err)
		}
		organisation.Members = []models.OrganisationMember{member}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to insert organisation or it's member: %w", err)
	}
	return &organisation, nil
}

func (s *S) DeleteOrganisation(userID string, organisationID int) error {
	// TODO:
	// Remove projects from orga
	// Set the organisation to deleted=true instead of removing it directly
	query := `DELETE FROM organisations WHERE owner_id = $1 AND id = $2`
	res, err := s.db.Exec(query, userID, organisationID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("unexpected number of rows affected, expected 1 but got %d for user %s", rowsAffected, o.OwnerID)
	}
	return nil
}

// SelectOrganisations returns a list of the user's Organisations
// User must be the owner of the organisation and also be a member of it.
func (s *S) SelectOrganisations(userID string) ([]Organisation, error) {
	var organisations []Organisation
	query := `
		SELECT o.id, o.name, o.owner_id = om.user_id as is_owner
		FROM organisations o
		JOIN organisation_members om ON o.id = om.organisation_id
		WHERE om.user_id = $1
	`
	if err := s.db.Select(&organisations, query, userID); err != nil {
		return nil, fmt.Errorf("failed to select organisations: %w", err)
	}
	return organisations, nil
}

// SelectOrganisation updates the Organisation struct with its members. Only if the User is the Owner of the Organisation
// The Organisation struct (o) must have the 'Name' and 'OwnerID' fields set before calling this function,
// as these fields are used to identify the specific organisation:
//
//   - o.Name: The name of the organisation for which member user IDs are to be retrieved.
//   - o.OwnerID: The ID of the owner of the organisation. Only the Owner can retrieve a list of members of the organisation
func (s *S) SelectOrganisation(orgID int, userID string) (*Organisation, error) {
	var organisation Organisation

	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		var exists int
		checkQuery := `
			SELECT 1
			FROM organisation_members
			WHERE organisation_id = $1 AND user_id = $2
		`
		if err := tx.QueryRow(checkQuery, orgID, userID).Scan(&exists); err != nil {
			return fmt.Errorf("failed to verify organisation membership: %w", err)
		}
		if exists == 0 {
			return errors.New("organisation not found or access denied")
		}

		orgQuery := `
			SELECT o.id, o.name, o.owner_id, u.user_id, u.email, u.username
			FROM organisations o
			INNER JOIN organisation_members om ON om.organisation_id = o.id
			INNER JOIN users u ON om.user_id = u.user_id
			WHERE o.id = $1
		`
		rows, err := tx.Query(orgQuery, orgID)
		if err != nil {
			return fmt.Errorf("failed to fetch organisation details: %w", err)
		}
		defer rows.Close()

		var organisationMember OrganisationMember
		var organisationMembers []OrganisationMember
		isOwner := false

		for rows.Next() {
			var ownerID string
			if err := rows.Scan(&organisation.ID, &organisation.Name, &ownerID, &organisationMember.UserID, &organisationMember.Email, &organisationMember.UserName); err != nil {
				return fmt.Errorf("failed to scan organisation data: %w", err)
			}
			organisationMembers = append(organisationMembers, organisationMember)
			if ownerID == userID {
				isOwner = true
			}
		}

		if organisation.ID == 0 {
			return errors.New("organisation not found")
		}

		organisation.Members = organisationMembers
		organisation.IsOwner = isOwner
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &organisation, nil
}

func (s *S) DeleteMember(ownerID, memberID, organisationID string) error {
	query := `
		DELETE FROM organisation_members
		WHERE user_id = $1 AND organisation_id = $3
		AND user_id <> $2
	`
	res, err := s.db.Exec(query, memberID, ownerID, organisationID)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("expected to delete 1 member from organisation '%s', but deleted %d", organisationID, rowsAffected)
	}
	return nil
}

func PutInvitation(ownerID, newMemberEmail string, organisationID int, invitationToken string, tx *sqlx.Tx) error {
	if isAlreadyMember := CheckIsMemberOfOrganisation(newMemberEmail, organisationID, tx); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of organisation: %d", newMemberEmail, organisationID)
	}

	query := `
	INSERT INTO organisation_invitations(organisation_id, email, invitation_token )
		SELECT id, $1, $4 FROM organisations WHERE owner_id = $2 AND id = $3;
	`
	res, err := tx.Exec(query, newMemberEmail, ownerID, organisationID, invitationToken)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organisations '%d', but added %d", organisationID, rem)
	}

	return nil
}

func (s *S) PutMember(newMemberID string, organisationID int) error {
	query := `
		INSERT INTO organisation_members(organisation_id, user_id)
		SELECT organisation_id, $1
		FROM organisation_invitations oi
		JOIN organisations o ON o.id = oi.organisation_id
		WHERE oi.user_id = $2 AND o.id = $3;
	`
	res, err := s.db.Exec(query, newMemberID, organisationID, organisationID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("expected to add 1 member to organisation '%d', but added %d", organisationID, affected)
	}
	return nil
}

func GetInvitations(userID string, organisationID int, tx *sqlx.Tx) ([]Invitation, error) {
	invites := make([]Invitation, 0)
	query := `SELECT oi.email, oi.organisation_id
				FROM organisation_invitations oi
				JOIN organisations o ON o.id = oi.organisation_id
				WHERE oi.organisation_id = $1
				ORDER BY oi.id DESC;
	`
	if err := tx.Select(&invites, query, organisationID); err != nil {
		return nil, err
	}
	return invites, nil
}

func WithdrawInvitation(userID, email string, organisationID int, tx *sqlx.Tx) error {
	query := `DELETE FROM organisation_invitations WHERE email=$1 AND organisation_id=$2`
	res, err := tx.Exec(query, email, organisationID)
	if err != nil {
		slog.Info("Error", "cant delete entry ", err)
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 invitation from organisations '%d', but deleted %d", organisationID, rem)
	}
	return nil
}

func CheckIsMemberOfOrganisation(userEmail string, organisationID int, tx *sqlx.Tx) bool {
	query := `SELECT 1 FROM organisation_members om
			  JOIN organisations o ON o.id = om.organisation_id
			  LEFT JOIN users u ON om.user_id = u.user_id
			  WHERE u.email = $1 AND o.id = $2;`
	isMemberOfSomeOrganisation := false
	_ = tx.Get(&isMemberOfSomeOrganisation, query, userEmail, organisationID)
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

func DeleteProject(userID string, organisationID int, upn string, tx *sqlx.Tx) error {
	query := `
		DELETE FROM projects_in_organisations
		WHERE organisation_id = $1 AND project_id = (
		SELECT id FROM projects
		WHERE user_id = $2 AND unique_name = $3
		);
	`
	res, err := tx.Exec(query, organisationID, userID, upn)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 project from organisations '%d', but deleted %d", organisationID, rem)
	}
	return nil
}
