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

// Response Object
type ChangeApplicationConfigurationResponse struct {
	// 应用ID。
	ApplicationId *string `json:"application_id,omitempty"`
	// 环境ID。
	EnvironmentId  *string                             `json:"environment_id,omitempty"`
	Configuration  *ApplicationListConfigConfiguration `json:"configuration,omitempty"`
	HttpStatusCode int                                 `json:"-"`
}

func (o ChangeApplicationConfigurationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeApplicationConfigurationResponse struct{}"
	}

	return strings.Join([]string{"ChangeApplicationConfigurationResponse", string(data)}, " ")
}
