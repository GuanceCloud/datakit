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
type DeleteNamespaceAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteNamespaceAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNamespaceAuthResponse struct{}"
	}

	return strings.Join([]string{"DeleteNamespaceAuthResponse", string(data)}, " ")
}
