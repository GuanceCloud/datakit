/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateScalingConfigResponse struct {
	// 伸缩配置ID
	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	HttpStatusCode         int     `json:"-"`
}

func (o CreateScalingConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateScalingConfigResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingConfigResponse", string(data)}, " ")
}
