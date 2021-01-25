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
type PublishMessageRequest struct {
	TopicUrn string                     `json:"topic_urn"`
	Body     *PublishMessageRequestBody `json:"body,omitempty"`
}

func (o PublishMessageRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublishMessageRequest struct{}"
	}

	return strings.Join([]string{"PublishMessageRequest", string(data)}, " ")
}
