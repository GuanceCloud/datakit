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
type DeleteApplicationConfigurationRequest struct {
	ApplicationId string `json:"application_id"`
	EnvironmentId string `json:"environment_id"`
}

func (o DeleteApplicationConfigurationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteApplicationConfigurationRequest struct{}"
	}

	return strings.Join([]string{"DeleteApplicationConfigurationRequest", string(data)}, " ")
}
