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
type ListSubscriptionsByTopicRequest struct {
	TopicUrn string `json:"topic_urn"`
	Offset   *int32 `json:"offset,omitempty"`
	Limit    *int32 `json:"limit,omitempty"`
}

func (o ListSubscriptionsByTopicRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubscriptionsByTopicRequest struct{}"
	}

	return strings.Join([]string{"ListSubscriptionsByTopicRequest", string(data)}, " ")
}
