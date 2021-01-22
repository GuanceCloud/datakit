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
type CreateTopicRequest struct {
	Body *CreateTopicRequestBody `json:"body,omitempty"`
}

func (o CreateTopicRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTopicRequest struct{}"
	}

	return strings.Join([]string{"CreateTopicRequest", string(data)}, " ")
}
