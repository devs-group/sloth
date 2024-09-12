package models

import "time"

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
	ID             int     `json:"id" db:"id"`
	UserID         int     `json:"user_id" db:"user_id"`
	Email          *string `json:"email,omitempty" db:"email"`
	UserName       *string `json:"username" db:"username"`
	OrganisationID int     `json:"organisation_id" db:"organisation_id"`
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
