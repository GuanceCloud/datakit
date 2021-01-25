/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CancelResourcesSubscriptionRequest struct {
	Body *UnsubscribeResourcesReq `json:"body,omitempty"`
}

func (o CancelResourcesSubscriptionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelResourcesSubscriptionRequest struct{}"
	}

	return strings.Join([]string{"CancelResourcesSubscriptionRequest", string(data)}, " ")
}
