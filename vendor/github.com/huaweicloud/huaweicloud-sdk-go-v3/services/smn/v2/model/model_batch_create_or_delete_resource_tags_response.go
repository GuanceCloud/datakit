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
type BatchCreateOrDeleteResourceTagsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchCreateOrDeleteResourceTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteResourceTagsResponse struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteResourceTagsResponse", string(data)}, " ")
}
