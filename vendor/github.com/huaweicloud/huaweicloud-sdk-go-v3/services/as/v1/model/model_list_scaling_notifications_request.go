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

// Request Object
type ListScalingNotificationsRequest struct {
	ScalingGroupId string `json:"scaling_group_id"`
}

func (o ListScalingNotificationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListScalingNotificationsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingNotificationsRequest", string(data)}, " ")
}
