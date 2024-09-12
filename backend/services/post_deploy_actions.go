package services

import (
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
)

type PostDeployAction struct {
	ID         int      `json:"id" db:"id"`
	ServiceId  int      `json:"service_id" db:"service_id"`
	Parameters []string `json:"parameters" db:"parameters"`
	Shell      string   `json:"shell" db:"shell"`
	Command    string   `json:"command" db:"command"`
}

type PostDeployActionResult struct {
	ID         int    `json:"id" db:"id"`
	ServiceId  int    `json:"service_id" db:"service_id"`
	Parameters string `json:"parameters" db:"parameters"`
	Shell      string `json:"shell" db:"shell"`
	Command    string `json:"command" db:"command"`
}

func StorePostDeployAction(serviceId int, parameters, shell, command string, tx *sqlx.Tx) error {
	query := `INSERT INTO post_deploy_actions (service_id, parameters, shell, command) VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(query, serviceId, parameters, shell, command)
	if err != nil {
		slog.Error("Unable to Store post_deploy_actions: %v", err)
		return err
	}

	return nil
}

func GetPostDeployActionsByServiceId(serviceID int, tx *sqlx.Tx) ([]PostDeployAction, error) {

	query := `SELECT * FROM post_deploy_actions WHERE service_id = $1`

	var post_deploy_actions_result = make([]PostDeployActionResult, 0)
	var post_deploy_actions = make([]PostDeployAction, 0)

	err := tx.Select(&post_deploy_actions_result, query, serviceID)
	if err != nil {
		slog.Error("Unable to find post_deploy_actions: %v", err)
		return nil, err
	}

	for _, pda := range post_deploy_actions_result {
		post_deploy_actions = append(post_deploy_actions, PostDeployAction{ID: pda.ID, ServiceId: pda.ServiceId, Parameters: strings.Split(pda.Parameters, ","), Command: pda.Command, Shell: pda.Shell})
	}

	return post_deploy_actions, nil
}
