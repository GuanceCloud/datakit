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

// Request Object
type DeleteResourceTagRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceId   string `json:"resource_id"`
	Key          string `json:"key"`
}

func (o DeleteResourceTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteResourceTagRequest struct{}"
	}

	return strings.Join([]string{"DeleteResourceTagRequest", string(data)}, " ")
}
