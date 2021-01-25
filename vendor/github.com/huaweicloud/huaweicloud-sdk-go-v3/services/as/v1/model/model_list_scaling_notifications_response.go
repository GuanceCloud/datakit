/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListScalingNotificationsResponse struct {
	// 伸缩组通知列表。
	Topics         *[]Topics `json:"topics,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListScalingNotificationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListScalingNotificationsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingNotificationsResponse", string(data)}, " ")
}
