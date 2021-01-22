/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ModifyInstanceConfigurationResponse struct {
	// 实例是否需要重启。  - “true”需要重启。 - “false”不需要重启。
	RestartRequired *bool `json:"restart_required,omitempty"`
	HttpStatusCode  int   `json:"-"`
}

func (o ModifyInstanceConfigurationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyInstanceConfigurationResponse struct{}"
	}

	return strings.Join([]string{"ModifyInstanceConfigurationResponse", string(data)}, " ")
}
