/*
 * Live
 *
 * 数据分析服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type PeakBandwidthData struct {
	// 带宽峰值，单位为bps。
	Value *int64 `json:"value,omitempty"`
	// 播放域名。
	Domain *string `json:"domain,omitempty"`
}

func (o PeakBandwidthData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PeakBandwidthData struct{}"
	}

	return strings.Join([]string{"PeakBandwidthData", string(data)}, " ")
}
