/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteNamespacesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteNamespacesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNamespacesResponse struct{}"
	}

	return strings.Join([]string{"DeleteNamespacesResponse", string(data)}, " ")
}
