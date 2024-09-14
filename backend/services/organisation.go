package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/models"
)

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
		return fmt.Errorf("failed to execute delete query for user with id %s and orga id %d: %w", userID, organisationID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("unexpected number of rows affected, expected 1 but got %d for user %s", rowsAffected, userID)
	}
	return nil
}

// SelectOrganisations returns a list of the user's Organisations
// User must be the owner of the organisation and also be a member of it.
func (s *S) SelectOrganisations(userID string) ([]models.Organisation, error) {
	var organisations []models.Organisation
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

func (s *S) SelectOrganisation(orgID int, userID string) (*models.Organisation, error) {
	var organisation models.Organisation

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

		var organisationMember models.OrganisationMember
		var organisationMembers []models.OrganisationMember
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

func (s *S) SaveInvitation(ownerID, newMemberEmail string, organisationID int, invitationToken string) error {
	if isAlreadyMember := s.CheckIsMemberOfOrganisation(newMemberEmail, organisationID); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of organisation: %d", newMemberEmail, organisationID)
	}
	query := `
	INSERT INTO organisation_invitations(organisation_id, email, invitation_token)
		SELECT id, $1, $4 FROM organisations WHERE owner_id = $2 AND id = $3;
	`
	res, err := s.db.Exec(query, newMemberEmail, ownerID, organisationID, invitationToken)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organisation with id '%d', but added %d", organisationID, rem)
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

func (s *S) GetInvitations(userID string, organisationID int) ([]models.Invitation, error) {
	invites := make([]models.Invitation, 0)
	query := `SELECT oi.email, oi.organisation_id
				FROM organisation_invitations oi
				JOIN organisations o ON o.id = oi.organisation_id
				WHERE oi.organisation_id = $1
				ORDER BY oi.id DESC;
	`
	if err := s.db.Select(&invites, query, organisationID); err != nil {
		return nil, fmt.Errorf("unable to query invitations", "err", err)
	}
	return invites, nil
}

func (s *S) WithdrawInvitation(userID, email string, organisationID int) error {
	query := `DELETE FROM organisation_invitations WHERE email=$1 AND organisation_id=$2`
	res, err := s.db.Exec(query, email, organisationID)
	if err != nil {
		return fmt.Errorf("unable to delete invitation: %w", err)
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 invitation from organisations '%d', but deleted %d", organisationID, rem)
	}
	return nil
}

func (s *S) CheckIsMemberOfOrganisation(userEmail string, organisationID int) bool {
	query := `SELECT 1 FROM organisation_members om
			  JOIN organisations o ON o.id = om.organisation_id
			  LEFT JOIN users u ON om.user_id = u.user_id
			  WHERE u.email = $1 AND o.id = $2;`
	isMemberOfSomeOrganisation := false
	_ = s.db.Get(&isMemberOfSomeOrganisation, query, userEmail, organisationID)
	return isMemberOfSomeOrganisation
}

func (s *S) GetInvitation(email, token string) (*models.Invitation, error) {
	var invitation models.Invitation
	q := `SELECT oi.email, o.name
	FROM organisation_invitations oi
	JOIN organisations o ON o.id = oi.organisation_id
	WHERE oi.email=$1 AND oi.invitation_token=$2`
	err := s.db.Get(&invitation, q, email, token)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (s *S) AcceptInvitation(userID, email, token string) (bool, error) {
	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		var accept models.AcceptInvite
		q := `SELECT timestamp, organisation_id FROM organisation_invitations WHERE email = $1 AND invitation_token = $2`
		err := tx.Get(&accept, q, email, token)
		if err != nil {
			return fmt.Errorf("unable to find invitation by email %s and token %s: %w", email, token, err)
		}
		q = `DELETE FROM organisation_invitations WHERE email=$1 AND invitation_token=$2 RETURNING organisation_id`
		err = tx.Get(&accept, q, email, token)
		if err != nil {
			return fmt.Errorf("unable to delete invitation by email %s and token %s: %w", email, token, err)
		}
		if time.Since(accept.TimeStamp) > config.EmailInvitationMaxValid {
			return fmt.Errorf("can't accept invitation, invitation too old. Timestamp %s", accept.TimeStamp)
		}
		q = `INSERT INTO organisation_members ( organisation_id, user_id ) VALUES ( $1, $2 )`
		res, err := tx.Exec(q, accept.OrganisationID, userID)
		if err != nil {
			return fmt.Errorf("unable to insert organisation member: %w", err)
		}
		ins, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if ins != 1 {
			return fmt.Errorf("unable to insert multiple users to organisations")
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *S) GetProjectsByOrganisationID(userID string, organisationID int) ([]models.OrganisationProjects, error) {
	projects := make([]models.OrganisationProjects, 0)
	q := `SELECT DISTINCT p.unique_name, p.name, p.id
	FROM projects p
	JOIN projects_in_organisations pio ON p.id = pio.project_id
	JOIN organisations o ON pio.organisation_id = o.id
	JOIN organisation_members om ON o.id = om.organisation_id
	WHERE o.id = $1 AND om.user_id = $2;
    `
	err := s.db.Select(&projects, q, organisationID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return projects, nil
		}
		return nil, err
	}
	return projects, nil
}

func (s *S) AddProjectToOrganisationByUPN(userID string, organisationID int, upn string) error {
	q := `
	INSERT INTO projects_in_organisations (project_id, organisation_id)
		SELECT p.id, $1
		FROM projects p
		WHERE p.unique_name = $2 AND p.user_id = $3 AND EXISTS (
    		SELECT 1 FROM organisations WHERE id = $1
		)
	`
	_, err := s.db.Exec(q, organisationID, upn, userID)
	if err != nil {
		return errors.Join(err, fmt.Errorf("can't add project your not owner or this project does not exist"))
	}
	return nil
}

func (s *S) DeleteProject(userID string, organisationID int, upn string) error {
	q := `
		DELETE FROM projects_in_organisations
		WHERE organisation_id = $1 AND project_id = (
			SELECT id FROM projects
			WHERE user_id = $2 AND unique_name = $3
		);
	`
	res, err := s.db.Exec(q, organisationID, userID, upn)
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
