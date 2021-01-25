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

type AdjustToIndirectPartnerReq struct {
	// |参数名称：合作伙伴关联的二级经销商伙伴ID。| |参数约束及描述：必填，最大长度64，合作伙伴关联的二级经销商伙伴ID。|
	IndirectPartnerId string `json:"indirect_partner_id"`
	// |参数名称：授信金额。单位为元不能为负数，精确到小数点后两位。| |参数的约束及描述：授信金额。单位为元不能为负数，精确到小数点后两位。|
	Amount float32 `json:"amount"`
}

func (o AdjustToIndirectPartnerReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AdjustToIndirectPartnerReq struct{}"
	}

	return strings.Join([]string{"AdjustToIndirectPartnerReq", string(data)}, " ")
}
