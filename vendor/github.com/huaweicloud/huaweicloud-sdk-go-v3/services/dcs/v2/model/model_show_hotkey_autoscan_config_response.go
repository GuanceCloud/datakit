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

// Response Object
type ShowHotkeyAutoscanConfigResponse struct {
	// 实例ID
	InstanceId *string `json:"instance_id,omitempty"`
	// 是否开启自动分析
	EnableAutoScan *bool `json:"enable_auto_scan,omitempty"`
	// 每日分析时间，时间格式为21:00
	ScheduleAt *[]string `json:"schedule_at,omitempty"`
	// 配置更新时间，时间格式为2020-06-15T02:21:18.669Z
	UpdatedAt      *string `json:"updated_at,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowHotkeyAutoscanConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowHotkeyAutoscanConfigResponse struct{}"
	}

	return strings.Join([]string{"ShowHotkeyAutoscanConfigResponse", string(data)}, " ")
}
