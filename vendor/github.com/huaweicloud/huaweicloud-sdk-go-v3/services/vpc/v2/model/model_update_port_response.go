/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdatePortResponse struct {
	Port           *Port `json:"port,omitempty"`
	HttpStatusCode int   `json:"-"`
}

func (o UpdatePortResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePortResponse struct{}"
	}

	return strings.Join([]string{"UpdatePortResponse", string(data)}, " ")
}
