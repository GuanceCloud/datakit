/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 扩容实例磁盘时必填。
type EnlargeVolume struct {
	// 扩容到该参数指定的大小，单位为GB。
	Size int32 `json:"size"`
	// 变更包周期实例的规格时可指定，表示是否自动从客户的账户中支付。
	IsAutoPay *bool `json:"is_auto_pay,omitempty"`
}

func (o EnlargeVolume) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnlargeVolume struct{}"
	}

	return strings.Join([]string{"EnlargeVolume", string(data)}, " ")
}
