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
type DeleteTopicAttributesRequest struct {
	TopicUrn string `json:"topic_urn"`
}

func (o DeleteTopicAttributesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTopicAttributesRequest struct{}"
	}

	return strings.Join([]string{"DeleteTopicAttributesRequest", string(data)}, " ")
}
