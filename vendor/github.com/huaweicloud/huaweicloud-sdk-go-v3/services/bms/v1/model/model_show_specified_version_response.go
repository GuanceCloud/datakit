/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowSpecifiedVersionResponse struct {
	Version        *Versions `json:"version,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ShowSpecifiedVersionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSpecifiedVersionResponse struct{}"
	}

	return strings.Join([]string{"ShowSpecifiedVersionResponse", string(data)}, " ")
}
