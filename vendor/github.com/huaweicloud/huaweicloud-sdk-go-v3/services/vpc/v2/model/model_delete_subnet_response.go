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
type DeleteSubnetResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteSubnetResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSubnetResponse struct{}"
	}

	return strings.Join([]string{"DeleteSubnetResponse", string(data)}, " ")
}
