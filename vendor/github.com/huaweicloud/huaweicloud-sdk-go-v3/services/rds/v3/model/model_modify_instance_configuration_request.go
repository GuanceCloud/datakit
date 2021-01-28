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

// Request Object
type ModifyInstanceConfigurationRequest struct {
	XLanguage  *string                             `json:"X-Language,omitempty"`
	InstanceId string                              `json:"instance_id"`
	Body       *ModifyInstanceConfigurationRequest `json:"body,omitempty"`
}

func (o ModifyInstanceConfigurationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyInstanceConfigurationRequest struct{}"
	}

	return strings.Join([]string{"ModifyInstanceConfigurationRequest", string(data)}, " ")
}
