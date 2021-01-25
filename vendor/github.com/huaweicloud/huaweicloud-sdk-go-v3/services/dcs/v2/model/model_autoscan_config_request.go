/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 设置自动分析配置返回体
type AutoscanConfigRequest struct {
	// 是否开启自动分析
	EnableAutoScan *bool `json:"enable_auto_scan,omitempty"`
	// 每日分析时间，时间格式为21:00，时间为UTC时间
	ScheduleAt *[]string `json:"schedule_at,omitempty"`
}

func (o AutoscanConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AutoscanConfigRequest struct{}"
	}

	return strings.Join([]string{"AutoscanConfigRequest", string(data)}, " ")
}
