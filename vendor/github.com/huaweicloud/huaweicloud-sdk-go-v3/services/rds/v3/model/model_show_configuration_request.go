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
type ShowConfigurationRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	ConfigId  string  `json:"config_id"`
}

func (o ShowConfigurationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowConfigurationRequest struct{}"
	}

	return strings.Join([]string{"ShowConfigurationRequest", string(data)}, " ")
}
