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
type CreateResourceTagRequest struct {
	ResourceType string                        `json:"resource_type"`
	ResourceId   string                        `json:"resource_id"`
	Body         *CreateResourceTagRequestBody `json:"body,omitempty"`
}

func (o CreateResourceTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateResourceTagRequest struct{}"
	}

	return strings.Join([]string{"CreateResourceTagRequest", string(data)}, " ")
}
