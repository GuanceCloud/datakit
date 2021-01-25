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

type V2BandwidthData struct {
	// 带宽值，单位为bps。
	Value *int64 `json:"value,omitempty"`
	// 采样时间。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ。
	Time *string `json:"time,omitempty"`
}

func (o V2BandwidthData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V2BandwidthData struct{}"
	}

	return strings.Join([]string{"V2BandwidthData", string(data)}, " ")
}
