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
type AddSubscriptionRequest struct {
	TopicUrn string                      `json:"topic_urn"`
	Body     *AddSubscriptionRequestBody `json:"body,omitempty"`
}

func (o AddSubscriptionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddSubscriptionRequest struct{}"
	}

	return strings.Join([]string{"AddSubscriptionRequest", string(data)}, " ")
}
