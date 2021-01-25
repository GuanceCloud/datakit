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

type AdjustAccountReq struct {
	// |参数名称：合作伙伴关联的客户的客户ID。| |参数约束及描述：合作伙伴关联的客户的客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：授信金额。单位为元不能为负数，精确到小数点后两位。| |参数的约束及描述：授信金额。单位为元不能为负数，精确到小数点后两位。|
	Amount float32 `json:"amount"`
	// |参数名称：二级经销商ID。| |参数约束及描述：二级经销商ID，如果一级经销商要给二级经销商的子客户设置折扣，需要携带这个字段。|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o AdjustAccountReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AdjustAccountReq struct{}"
	}

	return strings.Join([]string{"AdjustAccountReq", string(data)}, " ")
}
