package repository

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
)

type Organization struct {
	Name    string   `json:"organization_name" db:"name" binding:"required"`
	OwnerID string   `json:"-" db:"owner_id"`
	IsOwner bool     `json:"is_owner" db:"is_owner"`
	Members []string `json:"members"`
}

type Invitation struct {
	Email            string `json:"email" db:"email" binding:"required"`
	OrganizationName string `json:"organization_name" db:"name" binding:"required"`
}

type AcceptInvite struct {
	OrganizationID int       `json:"organization_id" db:"organization_id"`
	TimeStamp      time.Time `json:"timestamp" db:"timestamp"`
}

type OrganizationProjects struct {
	UnqiueName  string `json:"upn" db:"unique_name"`
	ProjectName string `json:"name" db:"name"`
}

func (o *Organization) CreateOrganization(tx *sqlx.Tx) error {
	var oID int
	query := `INSERT INTO organizations( name, owner_id ) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&oID, query, o.Name, o.OwnerID); err != nil {
		return err
	}

	var mID int
	query = `INSERT INTO organization_members (organization_id, user_id) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&mID, query, oID, o.OwnerID); err != nil {
		return err
	}

	return nil
}

// This function returns a list of the user's Organizations
// He must be the owner of the group and also be a member of it.
func SelectOrganizations(userID string, tx *sqlx.Tx) ([]Organization, error) {
	organizations := make([]Organization, 0)
	query := `SELECT o.name, 
					 o.owner_id = om.user_id as is_owner
				FROM organizations o
				JOIN organization_members om ON o.id = om.organization_id
				WHERE om.user_id = $1;
	`
	err := tx.Select(&organizations, query, userID)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

// This functions updates the Organization struct with its members. Only if the User is the Owner of the Organization
// The Organization struct (o) must have the 'Name' and 'OwnerID' fields set before calling this function,
// as these fields are used to identify the specific group:
//
//   - o.Name: The name of the group for which member user IDs are to be retrieved.
//   - o.OwnerID: The ID of the owner of the group. Only the Owner can retrieve a list of members of the group
func (g *Organization) SelectOrganization(tx *sqlx.Tx) error {
	members := make([]string, 0)
	query := `SELECT om.user_id
				FROM organization_members om
				INNER JOIN organizations o ON om.organization_id = o.id
				WHERE o.name = $1 AND o.owner_id = $2;
	`
	err := tx.Select(&members, query, g.Name, g.OwnerID)
	g.Members = members
	if err != nil {
		return err
	}

	// The Owner is also a Member so isOwner will always greater than 0
	// otherwise he's not the owner if the group
	if len(g.Members) > 0 {
		g.IsOwner = true
	}
	return nil
}

func (g *Organization) DeleteOrganization(tx *sqlx.Tx) error {
	// TODO: Remove project's from group
	query := `DELETE FROM organizations WHERE owner_id = $1 AND name = $2`
	res, err := tx.Exec(query, g.OwnerID, g.Name)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("found different amounts of organizations for user %s", g.OwnerID)
	}
	return nil
}

func DeleteMember(ownerID, memberID, organizationName string, tx *sqlx.Tx) error {
	query := `
		DELETE FROM organization_members
		WHERE user_id = $1 AND organization_id = (
		SELECT id FROM organizations
		WHERE owner_id = $2 AND name = $3
		) AND user_id <> $2;
	`
	res, err := tx.Exec(query, memberID, ownerID, organizationName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 member from organizations '%s', but deleted %d", organizationName, rem)
	}

	return nil
}

func PutInvitation(ownerID, newMemberID, organizationName, invitationToken string, tx *sqlx.Tx) error {
	if isAlreadyMember := CheckIsMemberOfOrganization(newMemberID, organizationName, tx); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of organization: %s", newMemberID, organizationName)
	}

	query := `
	INSERT INTO organization_invitations(organization_id, email, invitation_token )
		SELECT id, $1, $4 FROM organizations WHERE owner_id = $2 AND name = $3;
	`
	res, err := tx.Exec(query, newMemberID, ownerID, organizationName, invitationToken)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organizations '%s', but added %d", organizationName, rem)
	}

	return nil
}

func PutMember(newMemberID, organizationName string, tx *sqlx.Tx) error {
	query := `
    INSERT INTO organization_members(organization_id, user_id)
    	SELECT organization_id, $1 FROM organization_invitations oi
		JOIN organizations o ON o.id = oi.organization_id WHERE oi.user_id = $2 AND o.name = $3;
    `
	res, err := tx.Exec(query, newMemberID, organizationName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to organizations '%s', but added %d", organizationName, rem)
	}

	return nil
}

func GetInvitations(userID string, tx *sqlx.Tx) ([]Invitation, error) {
	invites := make([]Invitation, 0)
	query := `SELECT oi.email, o.name 
				FROM organization_invitations oi 
				JOIN organizations o ON o.id = oi.organization_id
				WHERE oi.email = $1
				ORDER BY oi.id DESC;
	`
	if err := tx.Select(&invites, query, userID); err != nil {
		return nil, err
	}
	return invites, nil
}

func CheckIsMemberOfOrganization(userID, organizationName string, tx *sqlx.Tx) bool {
	query := `SELECT 1 FROM organization_members om 
			  JOIN organizations o ON o.id = om.organization_id  WHERE user_id = $1 AND o.name = $2;`
	isMemberOfSomeOrganization := false
	_ = tx.Get(&isMemberOfSomeOrganization, query, userID, organizationName)
	return isMemberOfSomeOrganization
}

func GetInvitation(email, token string, tx *sqlx.Tx) (*Invitation, error) {
	query := `SELECT oi.email, o.name 
	FROM organization_invitations oi 
	JOIN organizations o ON o.id = oi.organization_id 
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

	query := `SELECT timestamp, organization_id FROM organization_invitations WHERE email = $1 AND invitation_token = $2`
	err := tx.Get(&accept, query, email, token)
	if err != nil {
		slog.Info("Error", "can't find invitation", err)
		return false, err
	}

	query = `DELETE FROM organization_invitations WHERE email=$1 AND invitation_token=$2 RETURNING organization_id`
	err = tx.Get(&accept, query, email, token)
	if err != nil {
		slog.Info("Error", "cant delete entry ", err)
		return false, err
	}

	if time.Since(accept.TimeStamp) > config.EmailInvitationMaxValid {
		slog.Info("Error", "The invitation is too old", accept)
		return false, fmt.Errorf("can't accept invitation, invitation too old")
	}

	query = `INSERT INTO organization_members ( organization_id, user_id ) VALUES ( $1, $2 )`
	res, err := tx.Exec(query, accept.OrganizationID, userID)
	if err != nil {
		slog.Info("Error", "can't insert new member to organization", err)
		return false, err
	}

	ins, err := res.RowsAffected()
	if err != nil {
		slog.Info("Error", "cant fetch inserted rows", err)
		return false, err
	}

	if ins != 1 {
		slog.Info("Error", "inserted multiple users to organization", "restricted")
		return false, fmt.Errorf("inserted multiple users to organization")
	}

	return true, nil
}

func GetProjectsByOrganizationName(userID, organizationName string, tx *sqlx.Tx) ([]OrganizationProjects, error) {
	projects := make([]OrganizationProjects, 0)
	query := `SELECT p.unique_name, p.name
	FROM projects p
	JOIN projects_in_organizations pio ON p.id = pio.project_id
	JOIN organizations o ON pio.organization_id = o.id
	JOIN organization_members om ON o.id = om.organization_id
	WHERE o.name = $1 AND om.user_id = $2;
	`

	err := tx.Select(&projects, query, organizationName, userID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func AddOrganizationProjectByUPN(userID, organizationName, upn string, tx *sqlx.Tx) (bool, error) {
	query := `
	INSERT INTO projects_in_organizations (project_id, organization_id)
	SELECT p.id, org.id
	FROM projects p, (SELECT id FROM organizations WHERE name = $1) as org
	WHERE p.unique_name = $2 AND p.user_id = $3	
	`
	_, err := tx.Exec(query, organizationName, upn, userID)
	if err != nil {
		return false, err
	}

	return true, nil
}
