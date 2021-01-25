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
type CreateNamespaceAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateNamespaceAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNamespaceAuthResponse struct{}"
	}

	return strings.Join([]string{"CreateNamespaceAuthResponse", string(data)}, " ")
}
