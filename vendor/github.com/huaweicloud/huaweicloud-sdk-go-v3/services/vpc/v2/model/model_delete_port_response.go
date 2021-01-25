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
type DeletePortResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeletePortResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePortResponse struct{}"
	}

	return strings.Join([]string{"DeletePortResponse", string(data)}, " ")
}
