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
type BatchCreateOrDeleteResourceTagsRequest struct {
	ResourceType string                                      `json:"resource_type"`
	ResourceId   string                                      `json:"resource_id"`
	Body         *BatchCreateOrDeleteResourceTagsRequestBody `json:"body,omitempty"`
}

func (o BatchCreateOrDeleteResourceTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteResourceTagsRequest struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteResourceTagsRequest", string(data)}, " ")
}
