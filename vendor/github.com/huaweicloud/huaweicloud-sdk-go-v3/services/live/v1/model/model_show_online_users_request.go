/*
 * Live
 *
 * 直播服务源站所有接口
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

// Request Object
type ShowOnlineUsersRequest struct {
	Domain     string           `json:"domain"`
	AppName    *string          `json:"app_name,omitempty"`
	StreamName *string          `json:"stream_name,omitempty"`
	StartTime  *sdktime.SdkTime `json:"start_time,omitempty"`
	EndTime    *sdktime.SdkTime `json:"end_time,omitempty"`
	Step       *int32           `json:"step,omitempty"`
}

func (o ShowOnlineUsersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowOnlineUsersRequest struct{}"
	}

	return strings.Join([]string{"ShowOnlineUsersRequest", string(data)}, " ")
}
