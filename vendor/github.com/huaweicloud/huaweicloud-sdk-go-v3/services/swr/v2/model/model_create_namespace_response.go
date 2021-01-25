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
type CreateNamespaceResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateNamespaceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNamespaceResponse struct{}"
	}

	return strings.Join([]string{"CreateNamespaceResponse", string(data)}, " ")
}
