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
type ShowEnvironmentDetailRequest struct {
	EnvironmentId string `json:"environment_id"`
}

func (o ShowEnvironmentDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowEnvironmentDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowEnvironmentDetailRequest", string(data)}, " ")
}
