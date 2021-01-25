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

type StreamForbiddenSetting struct {
	// 直播播放域名或推流域名
	Domain string `json:"domain"`
	// 流应用名称
	AppName string `json:"app_name"`
	// 流名称
	StreamName string `json:"stream_name"`
	// 恢复流时间，格式：yyyy-mm-ddThh:mm:ssZ，UTC时间，不指定则默认7天，最大禁推为90天
	ResumeTime *sdktime.SdkTime `json:"resume_time,omitempty"`
}

func (o StreamForbiddenSetting) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StreamForbiddenSetting struct{}"
	}

	return strings.Join([]string{"StreamForbiddenSetting", string(data)}, " ")
}
