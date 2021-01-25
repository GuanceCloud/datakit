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
type DeleteScalingNotificationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingNotificationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteScalingNotificationResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingNotificationResponse", string(data)}, " ")
}
