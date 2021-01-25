/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type StopMigrationTaskRequest struct {
	TaskId string `json:"task_id"`
}

func (o StopMigrationTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StopMigrationTaskRequest struct{}"
	}

	return strings.Join([]string{"StopMigrationTaskRequest", string(data)}, " ")
}
