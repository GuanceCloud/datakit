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
type ShowLifeCycleHookRequest struct {
	ScalingGroupId    string `json:"scaling_group_id"`
	LifecycleHookName string `json:"lifecycle_hook_name"`
}

func (o ShowLifeCycleHookRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowLifeCycleHookRequest struct{}"
	}

	return strings.Join([]string{"ShowLifeCycleHookRequest", string(data)}, " ")
}
