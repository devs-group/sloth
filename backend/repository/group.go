package repository

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Group struct {
	Name    string   `json:"group_name" db:"name" binding:"required"`
	OwnerID string   `json:"-" db:"owner_id"`
	IsOwner bool     `json:"is_owner" db:"is_owner"`
	Members []string `json:"members"`
}

type Invitation struct {
	Email     string `json:"email" db:"email" binding:"required"`
	GroupName string `json:"group_name" db:"name" binding:"required"`
}

func (o *Group) CreateGroup(tx *sqlx.Tx) error {
	var oID int
	query := `INSERT INTO groups( name, owner_id ) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&oID, query, o.Name, o.OwnerID); err != nil {
		return err
	}

	var mID int
	query = `INSERT INTO group_members (group_id, user_id) VALUES ( $1, $2 ) RETURNING id`
	if err := tx.Get(&mID, query, oID, o.OwnerID); err != nil {
		return err
	}

	return nil
}

// This function returns a list of the user's Groups
// He must be the owner of the group and also be a member of it.
func SelectGroups(userID string, tx *sqlx.Tx) ([]Group, error) {
	groups := make([]Group, 0)
	query := `SELECT g.name, 
					 g.owner_id = gm.user_id as is_owner
				FROM groups g
				JOIN group_members gm ON g.id = gm.group_id
				WHERE gm.user_id = $1;
	`
	err := tx.Select(&groups, query, userID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// This functions updates the Group struct with its members. Only if the User is the Owner of the Group
// The Group struct (o) must have the 'Name' and 'OwnerID' fields set before calling this function,
// as these fields are used to identify the specific group:
//
//   - o.Name: The name of the group for which member user IDs are to be retrieved.
//   - o.OwnerID: The ID of the owner of the group. Only the Owner can retrieve a list of members of the group
func (g *Group) SelectGroup(tx *sqlx.Tx) error {
	members := make([]string, 0)
	query := `SELECT gm.user_id
				FROM group_members gm
				INNER JOIN groups g ON gm.group_id = g.id
				WHERE g.name = $1 AND g.owner_id = $2;
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

func (g *Group) DeleteGroup(tx *sqlx.Tx) error {
	// TODO: Remove project's from group
	query := `DELETE FROM groups WHERE owner_id = $1 AND name = $2`
	res, err := tx.Exec(query, g.OwnerID, g.Name)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("found different amounts of groups for user %s", g.OwnerID)
	}
	return nil
}

func DeleteMember(ownerID, memberID, groupName string, tx *sqlx.Tx) error {
	query := `
		DELETE FROM group_members
		WHERE user_id = $1 AND group_id = (
		SELECT id FROM groups
		WHERE owner_id = $2 AND name = $3
		) AND user_id <> $2;
	`
	res, err := tx.Exec(query, memberID, ownerID, groupName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to delete 1 member from group '%s', but deleted %d", groupName, rem)
	}

	return nil
}

func PutInvitation(ownerID, newMemberID, groupName, invitationToken string, tx *sqlx.Tx) error {
	if isAlreadyMember := CheckIsMemberOfGroup(newMemberID, groupName, tx); isAlreadyMember {
		return fmt.Errorf("user: %s is already member of group: %s", newMemberID, groupName)
	}

	query := `
	INSERT INTO group_invitations(group_id, email, invitation_token )
		SELECT id, $1, $4 FROM groups WHERE owner_id = $2 AND name = $3;
	`
	res, err := tx.Exec(query, newMemberID, ownerID, groupName, invitationToken)
	if err != nil {
		return err
	}
	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to group '%s', but added %d", groupName, rem)
	}

	return nil
}

func PutMember(newMemberID, groupName string, tx *sqlx.Tx) error {
	query := `
    INSERT INTO group_members(group_id, user_id)
    	SELECT group_id, $1 FROM groups_invitations gi
		JOIN groups g ON g.id = gi.group_id WHERE gi.user_id = $2 AND g.name = $3;
    `
	res, err := tx.Exec(query, newMemberID, groupName)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("expected to add 1 member to group '%s', but added %d", groupName, rem)
	}

	return nil
}

func GetInvitations(userID string, tx *sqlx.Tx) ([]Invitation, error) {
	invites := make([]Invitation, 0)
	query := `SELECT gi.email, g.name 
				FROM group_invitations gi 
				JOIN groups g ON g.id = gi.group_id
				WHERE gi.email = $1
				ORDER BY gi.id DESC;
	`
	if err := tx.Select(&invites, query, userID); err != nil {
		return nil, err
	}
	return invites, nil
}

func CheckIsMemberOfGroup(userID, groupName string, tx *sqlx.Tx) bool {
	query := `SELECT 1 FROM group_members gm JOIN groups g ON g.id = gm.group_id  WHERE user_id = $1 AND g.name = $2;`
	isMemberOfSomeGroup := false
	_ = tx.Get(&isMemberOfSomeGroup, query, userID, groupName)
	return isMemberOfSomeGroup
}

func GetInvitation(email, token string, tx *sqlx.Tx) (*Invitation, error) {
	query := `SELECT gi.email, g.name FROM group_invitations gi JOIN groups g ON g.id = gi.group_id WHERE gi.email=$1 AND gi.invitation_token=$2`
	var invitation Invitation
	err := tx.Get(&invitation, query, email, token)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func AcceptInvitation(userID, email, token string, tx *sqlx.Tx) (bool, error) {
	type AcceptInvite struct {
		GroupID int `db:"group_id"`
	}
	var accept AcceptInvite
	query := `DELETE FROM group_invitations WHERE email=$1 AND invitation_token=$2 RETURNING group_id`
	err := tx.Get(&accept, query, email, token)
	if err != nil {
		slog.Info("Delete", "del", err)
		return false, err
	}

	query = `INSERT INTO group_members ( group_id, user_id ) VALUES ( $1, $2 )`
	res, err := tx.Exec(query, accept.GroupID, userID)
	if err != nil {
		slog.Info("insert", "cant", err)
		return false, err
	}
	ins, err := res.RowsAffected()
	if err != nil {
		slog.Info("info", "rows", err)
		return false, err
	}

	if ins != 1 {
		slog.Info("Info", "inf", "multuple inserts")
		return false, fmt.Errorf("Multiple inserts!")
	}

	return true, nil
}

type GroupProjects struct {
	UnqiueName  string `json:"upn" db:"unique_name"`
	ProjectName string `json:"name" db:"name"`
}

func GetProjectsByGroupName(userID, groupName string, tx *sqlx.Tx) ([]GroupProjects, error) {
	projects := make([]GroupProjects, 0)
	query := `SELECT p.unique_name, p.name FROM projects p JOIN groups g ON p.group_id = g.id JOIN group_members gm ON g.id = gm.group_id WHERE g.name = $1 AND gm.user_id = $2`
	err := tx.Select(&projects, query, groupName, userID)
	if err != nil {
		return nil, err
	}
	slog.Info("Groups", "p", projects)
	return projects, nil
}

func AddGroupProjectByUPN(userID, groupName, upn string, tx *sqlx.Tx) (bool, error) {
	query := `UPDATE projects
	SET group_id = (
	  SELECT g.id
	  FROM groups g
	  JOIN group_members gm ON g.id = gm.group_id
	  WHERE g.name = $1
	  LIMIT 1
	)
	WHERE unique_name = $2 AND user_id = $3;
	`
	res, err := tx.Exec(query, groupName, upn, userID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, fmt.Errorf("Updated multuple rows ")
	}

	return true, nil
}
