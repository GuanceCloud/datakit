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
type DeleteTopicAttributeByNameRequest struct {
	TopicUrn string `json:"topic_urn"`
	Name     string `json:"name"`
}

func (o DeleteTopicAttributeByNameRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTopicAttributeByNameRequest struct{}"
	}

	return strings.Join([]string{"DeleteTopicAttributeByNameRequest", string(data)}, " ")
}
