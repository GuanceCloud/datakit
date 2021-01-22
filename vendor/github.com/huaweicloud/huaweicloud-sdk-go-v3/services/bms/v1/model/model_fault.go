/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

// fault字段数据结构说明
type Fault struct {
	// 故障信息
	Message *string `json:"message,omitempty"`
	// 故障code
	Code *int32 `json:"code,omitempty"`
	// 故障详情
	Details *string `json:"details,omitempty"`
	// 故障时间。时间戳格式为ISO 8601：YYYY-MM-DDTHH:MM:SSZ，例如：2019-05-22T03:30:52Z
	Created *sdktime.SdkTime `json:"created,omitempty"`
}

func (o Fault) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Fault struct{}"
	}

	return strings.Join([]string{"Fault", string(data)}, " ")
}
