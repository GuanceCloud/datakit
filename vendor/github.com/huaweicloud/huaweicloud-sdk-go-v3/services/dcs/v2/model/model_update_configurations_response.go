/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateConfigurationsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateConfigurationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateConfigurationsResponse struct{}"
	}

	return strings.Join([]string{"UpdateConfigurationsResponse", string(data)}, " ")
}
