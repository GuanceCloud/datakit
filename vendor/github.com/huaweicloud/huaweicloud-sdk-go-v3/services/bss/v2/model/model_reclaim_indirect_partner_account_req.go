/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ReclaimIndirectPartnerAccountReq struct {
	// |参数名称：精英服务商伙伴的ID。| |参数约束及描述：精英服务商伙伴的ID。|
	IndirectPartnerId string `json:"indirect_partner_id"`
	// |参数名称：拨款金额。单位为元。不能为负数，精确到小数点后两位。| |参数的约束及描述：拨款金额。单位为元。不能为负数，浮点数精度为:小数点后两位。|
	Amount float32 `json:"amount"`
}

func (o ReclaimIndirectPartnerAccountReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimIndirectPartnerAccountReq struct{}"
	}

	return strings.Join([]string{"ReclaimIndirectPartnerAccountReq", string(data)}, " ")
}
