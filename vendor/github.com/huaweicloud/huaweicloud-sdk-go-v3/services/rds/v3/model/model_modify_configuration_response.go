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
type ModifyConfigurationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ModifyConfigurationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyConfigurationResponse struct{}"
	}

	return strings.Join([]string{"ModifyConfigurationResponse", string(data)}, " ")
}
