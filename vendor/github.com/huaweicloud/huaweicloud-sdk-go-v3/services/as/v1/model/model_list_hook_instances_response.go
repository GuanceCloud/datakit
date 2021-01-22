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
type ListHookInstancesResponse struct {
	// 伸缩实例生命周期挂钩列表。
	InstanceHangingInfo *[]InstanceHangingInfos `json:"instance_hanging_info,omitempty"`
	HttpStatusCode      int                     `json:"-"`
}

func (o ListHookInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHookInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListHookInstancesResponse", string(data)}, " ")
}
