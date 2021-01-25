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
type ListTopicDetailsRequest struct {
	TopicUrn string `json:"topic_urn"`
}

func (o ListTopicDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicDetailsRequest struct{}"
	}

	return strings.Join([]string{"ListTopicDetailsRequest", string(data)}, " ")
}
