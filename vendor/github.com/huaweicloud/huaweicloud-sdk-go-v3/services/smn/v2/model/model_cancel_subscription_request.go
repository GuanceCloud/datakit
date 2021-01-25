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
type CancelSubscriptionRequest struct {
	SubscriptionUrn string `json:"subscription_urn"`
}

func (o CancelSubscriptionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelSubscriptionRequest struct{}"
	}

	return strings.Join([]string{"CancelSubscriptionRequest", string(data)}, " ")
}
