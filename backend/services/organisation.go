package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/models"
)

func (s *S) CreateOrganisation(o models.Organisation, userID string) (*models.Organisation, error) {
	var organisation models.Organisation
	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		query := `INSERT INTO organisations( name, is_default ) VALUES ( $1, $2 ) RETURNING *`
		if err := tx.Get(&organisation, query, o.Name, false); err != nil {
			return fmt.Errorf("unable to create organisation: %w", err)
		}

		var member models.OrganisationMember
		query = `INSERT INTO organisation_members (organisation_id, user_id, role) VALUES ( $1, $2, $3 ) RETURNING *`
		if err := tx.Get(&member, query, organisation.ID, userID, "owner"); err != nil {
			return fmt.Errorf("unable to add member to organisation: %w", err)
		}
		organisation.Members = []models.OrganisationMember{member}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to insert organisation or it's member: %w", err)
	}
	return &organisation, nil
}

func (s *S) UpdateOrganisation(o models.Organisation, organisationID int, userID string) error {
	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		query := `
			UPDATE organisations
			SET name=$1
			WHERE id = $2
			AND EXISTS (
				SELECT 1
				FROM organisation_members
				WHERE organisation_id = organisations.id
				  AND user_id = $3
				  AND role IN ('owner', 'admin')
		  	);
		`
		result, err := tx.Exec(query, o.Name, organisationID, userID)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return errors.New("no changes were made")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S) DeleteOrganisation(userID string, organisationID int) error {
	// TODO: Set the organisation to deleted=true instead of removing it directly
	query := `
		DELETE FROM organisations
	   	WHERE id = $1
	   	AND is_default = false
		AND EXISTS (
			SELECT 1 FROM organisation_members
			WHERE organisation_members.organisation_id = organisations.id
			AND organisation_members.user_id = $2
			AND organisation_members.role = 'owner'
		);
	`
	res, err := s.dbService.GetConn().Exec(query, organisationID, userID)
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
		SELECT o.id
		FROM organisations o
		JOIN organisation_members om ON o.id = om.organisation_id
		WHERE om.user_id = $1
	`
	rows, err := s.dbService.GetConn().Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users organisations: %w", err)
	}

	for rows.Next() {
		var organisationID int
		err = rows.Scan(&organisationID)
		if err != nil {
			return nil, err
		}
		organisation, err := s.GetOrganisation(organisationID, userID)
		if err != nil {
			return nil, err
		}

		organisations = append(organisations, *organisation)
	}

	return organisations, nil
}

func (s *S) GetOrganisation(orgID int, userID string) (*models.Organisation, error) {
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
			SELECT o.id, o.name, o.is_default, om.id, om.role, u.user_id, u.email, u.username
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
		userIDInt, err := strconv.Atoi(userID)
		currentRole := "member"

		for rows.Next() {
			if err := rows.Scan(&organisation.ID, &organisation.Name, &organisation.IsDefault, &organisationMember.ID, &organisationMember.Role, &organisationMember.UserID, &organisationMember.Email, &organisationMember.UserName); err != nil {
				return fmt.Errorf("failed to scan organisation data: %w", err)
			}
			organisationMembers = append(organisationMembers, organisationMember)
			if organisationMember.UserID == userIDInt {
				currentRole = organisationMember.Role
			}
		}

		if organisation.ID == 0 {
			return errors.New("organisation not found")
		}

		organisation.Members = organisationMembers
		organisation.CurrentRole = currentRole
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &organisation, nil
}

func (s *S) DeleteMember(userID, memberID, organisationID string) error {
	query := `
		DELETE FROM organisation_members
		WHERE id = $1
	    AND organisation_id = $2
		AND EXISTS (
			SELECT 1
			FROM organisation_members om
			WHERE $2 = om.organisation_id
			  AND om.user_id = $3 AND om.role IN ('owner', 'admin')
			  OR om.id = $1 AND om.user_id = $3 AND om.role IN ('member', 'admin')
		);
	`
	res, err := s.dbService.GetConn().Exec(query, memberID, organisationID, userID)
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

func (s *S) CreateOrganisationInvitation(newMemberEmail string, organisationID int, invitationToken string) error {
	if isAlreadyMember := s.CheckIsMemberOfOrganisation(newMemberEmail, organisationID); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of organisation: %d", newMemberEmail, organisationID)
	}
	validUntil := time.Now().Add(time.Hour * 24)
	query := `
		INSERT INTO organisation_invitations(email, invitation_token, valid_until, organisation_id)
		VALUES ($1, $2, $3, $4);
	`
	res, err := s.dbService.GetConn().Exec(
		query,
		newMemberEmail,
		invitationToken,
		validUntil,
		organisationID,
	)
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
		INSERT INTO organisation_members(organisation_id, user_id, role)
		SELECT organisation_id, $1, $2
		FROM organisation_invitations oi
		JOIN organisations o ON o.id = oi.organisation_id
		WHERE o.id = $2;
	`
	res, err := s.dbService.GetConn().Exec(query, newMemberID, "member", organisationID)
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

func (s *S) GetInvitations(organisationID int) ([]models.Invitation, error) {
	invites := make([]models.Invitation, 0)
	query := `SELECT oi.id, oi.email, oi.organisation_id, oi.valid_until
				FROM organisation_invitations oi
				JOIN organisations o ON o.id = oi.organisation_id
				WHERE oi.organisation_id = $1
				ORDER BY oi.id DESC;
	`
	if err := s.dbService.GetConn().Select(&invites, query, organisationID); err != nil {
		return nil, fmt.Errorf("unable to query invitations: %w", err)
	}
	return invites, nil
}

func (s *S) DeleteOrganisationInvitation(invitationID int, userID string) error {
	query := `
		DELETE FROM organisation_invitations
	   	WHERE id=$1
		AND EXISTS (
			SELECT 1
			FROM organisation_members
			WHERE organisation_id = organisation_invitations.organisation_id
			  AND user_id = $2
			  AND role IN ('owner', 'admin')
		);
  	`
	res, err := s.dbService.GetConn().Exec(query, invitationID, userID)
	if err != nil {
		return fmt.Errorf("unable to delete invitation: %w", err)
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("no change after deletion")
	}
	return nil
}

func (s *S) CheckIsMemberOfOrganisation(userEmail string, organisationID int) bool {
	query := `SELECT 1 FROM organisation_members om
			  JOIN organisations o ON o.id = om.organisation_id
			  LEFT JOIN users u ON om.user_id = u.user_id
			  WHERE u.email = $1 AND o.id = $2;`
	isMemberOfSomeOrganisation := false
	_ = s.dbService.GetConn().Get(&isMemberOfSomeOrganisation, query, userEmail, organisationID)
	return isMemberOfSomeOrganisation
}

func (s *S) CheckIsMemberAlreadyInvited(userEmail string, organisationID int) bool {
	query := `SELECT 1 FROM organisation_invitations oi
			  WHERE oi.email = $1 AND oi.id = $2 AND oi.valid_until > DATE();`
	isMemberAlreadyInvited := false
	_ = s.dbService.GetConn().Get(&isMemberAlreadyInvited, query, userEmail, organisationID)
	return isMemberAlreadyInvited
}

func (s *S) GetInvitation(email, token string) (*models.Invitation, error) {
	var invitation models.Invitation
	q := `SELECT oi.email, o.name
	FROM organisation_invitations oi
	JOIN organisations o ON o.id = oi.organisation_id
	WHERE oi.email=$1 AND oi.invitation_token=$2`
	err := s.dbService.GetConn().Get(&invitation, q, email, token)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (s *S) AcceptInvitation(userID, email, token string) (bool, error) {
	cfg := config.GetConfig()

	err := s.WithTransaction(func(tx *sqlx.Tx) error {
		var accept models.AcceptInvite
		q := `SELECT valid_until, organisation_id FROM organisation_invitations WHERE email = $1 AND invitation_token = $2`
		err := tx.Get(&accept, q, email, token)
		if err != nil {
			return fmt.Errorf("unable to find invitation by email %s and token %s: %w", email, token, err)
		}
		q = `DELETE FROM organisation_invitations WHERE email=$1 AND invitation_token=$2 RETURNING organisation_id`
		err = tx.Get(&accept, q, email, token)
		if err != nil {
			return fmt.Errorf("unable to delete invitation by email %s and token %s: %w", email, token, err)
		}
		if time.Since(accept.ValidUntil) > cfg.EmailInvitationMaxValid {
			return fmt.Errorf("can't accept invitation, invitation too old. Timestamp %s", accept.ValidUntil)
		}
		q = `INSERT INTO organisation_members ( organisation_id, user_id, role ) VALUES ( $1, $2, $3 )`
		res, err := tx.Exec(q, accept.OrganisationID, userID, "member")
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

func (s *S) GetProjectsByOrganisationID(organisationID int) ([]models.OrganisationProjects, error) {
	projects := make([]models.OrganisationProjects, 0)
	q := `SELECT DISTINCT p.unique_name, p.name, p.id
	FROM projects p
	WHERE p.organisation_id = $1;
    `
	err := s.dbService.GetConn().Select(&projects, q, organisationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return projects, nil
		}
		return nil, err
	}
	return projects, nil
}

func (s *S) AddProjectToOrganisationByUPN(organisationID int, upn string) error {
	q := `
		UPDATE projects
		SET organisation_id = $1
		WHERE unique_name = $2
	`
	_, err := s.dbService.GetConn().Exec(q, organisationID, upn)
	if err != nil {
		return errors.Join(err, fmt.Errorf("can't add project your not owner or this project does not exist"))
	}
	return nil
}

func (s *S) RemoveProjectFromOrganisation(organisationID int, upn string) error {
	q := `
		UPDATE projects
		SET organisation_id = NULL
		WHERE organisation_id = $1 AND unique_name = $2
	`
	res, err := s.dbService.GetConn().Exec(q, organisationID, upn)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to remove 1 project from organisation with ID '%d', but deleted %d", organisationID, rem)
	}
	return nil
}
