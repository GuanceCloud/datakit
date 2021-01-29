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
type DeleteConfigurationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteConfigurationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteConfigurationResponse struct{}"
	}

	return strings.Join([]string{"DeleteConfigurationResponse", string(data)}, " ")
}
