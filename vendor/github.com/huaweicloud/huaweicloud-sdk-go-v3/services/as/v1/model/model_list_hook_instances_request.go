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
type ListHookInstancesRequest struct {
	ScalingGroupId string  `json:"scaling_group_id"`
	InstanceId     *string `json:"instance_id,omitempty"`
}

func (o ListHookInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHookInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListHookInstancesRequest", string(data)}, " ")
}
