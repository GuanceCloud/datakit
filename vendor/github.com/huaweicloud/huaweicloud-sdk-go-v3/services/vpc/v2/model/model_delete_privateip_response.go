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
type DeletePrivateipResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeletePrivateipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePrivateipResponse struct{}"
	}

	return strings.Join([]string{"DeletePrivateipResponse", string(data)}, " ")
}
