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
type CreatePortResponse struct {
	Port           *Port `json:"port,omitempty"`
	HttpStatusCode int   `json:"-"`
}

func (o CreatePortResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePortResponse struct{}"
	}

	return strings.Join([]string{"CreatePortResponse", string(data)}, " ")
}
