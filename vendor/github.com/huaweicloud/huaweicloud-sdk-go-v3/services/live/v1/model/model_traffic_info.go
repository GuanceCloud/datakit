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

type TrafficInfo struct {
	// 采样周期内的总流量，单位：byte
	Traffic int32 `json:"traffic"`
	// 流量数据采样周期起始时刻，UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ
	Timestamp *sdktime.SdkTime `json:"timestamp"`
}

func (o TrafficInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TrafficInfo struct{}"
	}

	return strings.Join([]string{"TrafficInfo", string(data)}, " ")
}
