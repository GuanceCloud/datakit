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

type TrafficSummaryData struct {
	// 流量，单位为byte。
	Value *int64 `json:"value,omitempty"`
	// 域名。
	Domain *string `json:"domain,omitempty"`
}

func (o TrafficSummaryData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TrafficSummaryData struct{}"
	}

	return strings.Join([]string{"TrafficSummaryData", string(data)}, " ")
}
