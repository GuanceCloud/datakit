/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateResourceTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateResourceTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateResourceTagResponse struct{}"
	}

	return strings.Join([]string{"CreateResourceTagResponse", string(data)}, " ")
}
