package models

import "time"

type Organisation struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"organisationName" db:"name" binding:"required"`
	IsDefault bool      `json:"isDefault" db:"is_default"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	CurrentRole string               `json:"currentRole"`
	Members     []OrganisationMember `json:"members"`
}

type Invitation struct {
	ID             int       `json:"id" db:"id"`
	Email          string    `json:"email" db:"email" binding:"required"`
	OrganisationID int       `json:"organisationID" db:"organisation_id" binding:"required"`
	ValidUntil     time.Time `json:"validUntil" db:"valid_until"`
}

type AcceptInvite struct {
	OrganisationID int       `json:"organisation_id" db:"organisation_id"`
	ValidUntil     time.Time `json:"validUntil" db:"valid_until"`
}

type OrganisationProjects struct {
	UniqueName  string `json:"upn" db:"unique_name"`
	ProjectName string `json:"name" db:"name"`
	ID          int    `json:"id"   db:"id"`
}

type OrganisationMember struct {
	ID             int     `json:"id" db:"id"`
	Role           string  `json:"role" db:"role"`
	UserID         int     `json:"userID" db:"user_id"`
	OrganisationID int     `json:"organisationID" db:"organisation_id"`
	Email          *string `json:"email,omitempty" db:"email"`
	UserName       *string `json:"username" db:"username"`
}
