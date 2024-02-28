package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Group struct {
	Name    string   `json:"group_name" db:"name" binding:"required"`
	OwnerID string   `json:"-" db:"owner_id"`
	IsOwner bool     `json:"is_owner" db:"is_owner"`
	Members []string `json:"members"`
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

func PutMember(ownerID, memberID, groupName string, tx *sqlx.Tx) error {
	query := `
    INSERT INTO group_members(group_id, user_id)
    	SELECT id, $1 FROM groups WHERE owner_id = $2 AND name = $3;
    `
	res, err := tx.Exec(query, memberID, ownerID, groupName)
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

func CheckMemberOfGroup(userID, groupName string, tx *sqlx.Tx) bool {
	query := `SELECT 1 FROM group_members gm JOIN groups g ON g.id = gm.group_id  WHERE user_id = $1 AND g.name = $2;`
	isMemberOfSomeGroup := false
	_ = tx.Get(&isMemberOfSomeGroup, query, userID, groupName)
	return isMemberOfSomeGroup
}
