package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Group struct {
	Name    string `json:"group_name" db:"name" binding:"required"`
	OwnerID string `json:"-" db:"owner_id"`
	IsOwner bool   `json:"is_owner" db:"is_owner"`
	UserID  []int  `json:"users"`
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

func (o *Group) DeleteGroup(tx *sqlx.Tx) error {
	// TODO REMOVE PROJECTS FROM GROUP

	query := `DELETE FROM groups WHERE owner_id = $1 AND name = $2`
	res, err := tx.Exec(query, o.OwnerID, o.Name)
	if err != nil {
		return err
	}

	rem, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rem != 1 {
		return fmt.Errorf("found different amounts of groups for user %s", o.OwnerID)
	}
	return nil
}

func (o *Group) SelectGroup(tx *sqlx.Tx) error {
	return nil
}
