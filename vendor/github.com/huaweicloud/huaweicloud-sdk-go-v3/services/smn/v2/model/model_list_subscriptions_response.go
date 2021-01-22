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

// Response Object
type ListSubscriptionsResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 订阅者个数。
	SubscriptionCount *int32 `json:"subscription_count,omitempty"`
	// Subscription结构体。
	Subscriptions  *[]ListSubscriptionsItem `json:"subscriptions,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o ListSubscriptionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubscriptionsResponse struct{}"
	}

	return strings.Join([]string{"ListSubscriptionsResponse", string(data)}, " ")
}
