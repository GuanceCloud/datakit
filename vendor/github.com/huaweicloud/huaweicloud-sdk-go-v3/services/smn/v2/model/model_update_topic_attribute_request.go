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
type UpdateTopicAttributeRequest struct {
	TopicUrn string                           `json:"topic_urn"`
	Name     string                           `json:"name"`
	Body     *UpdateTopicAttributeRequestBody `json:"body,omitempty"`
}

func (o UpdateTopicAttributeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTopicAttributeRequest struct{}"
	}

	return strings.Join([]string{"UpdateTopicAttributeRequest", string(data)}, " ")
}
