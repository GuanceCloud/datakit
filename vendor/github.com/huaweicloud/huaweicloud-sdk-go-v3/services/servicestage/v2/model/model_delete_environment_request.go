/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteEnvironmentRequest struct {
	EnvironmentId string `json:"environment_id"`
}

func (o DeleteEnvironmentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteEnvironmentRequest struct{}"
	}

	return strings.Join([]string{"DeleteEnvironmentRequest", string(data)}, " ")
}
