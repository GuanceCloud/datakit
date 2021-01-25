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
type ShowApplicationConfigurationResponse struct {
	// 应用配置列表。
	Configuration  *[]ApplicationListConfigConfiguration1 `json:"configuration,omitempty"`
	HttpStatusCode int                                    `json:"-"`
}

func (o ShowApplicationConfigurationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowApplicationConfigurationResponse struct{}"
	}

	return strings.Join([]string{"ShowApplicationConfigurationResponse", string(data)}, " ")
}
