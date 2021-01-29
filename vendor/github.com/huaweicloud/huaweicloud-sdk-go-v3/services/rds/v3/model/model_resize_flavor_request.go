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

// 变更实例规格时必填。
type ResizeFlavorRequest struct {
	// 资源规格编码。
	SpecCode string `json:"spec_code"`
	// 变更包周期实例的规格时可指定，表示是否自动从客户的账户中支付。
	IsAutoPay *bool `json:"is_auto_pay,omitempty"`
}

func (o ResizeFlavorRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeFlavorRequest struct{}"
	}

	return strings.Join([]string{"ResizeFlavorRequest", string(data)}, " ")
}
