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

type AdjustCouponQuotasReq struct {
	// |参数名称：优惠券额度ID。| |参数约束及描述：优惠券额度ID。|
	QuotaId string `json:"quota_id"`
	// |参数名称：二级分销商伙伴id列表。最大100条| |参数约束以及描述：二级分销商伙伴id列表。最大100条|
	IndirectPartnerIds []string `json:"indirect_partner_ids"`
	// |参数名称：额度值。保留小数点后2位| |参数的约束及描述：额度值。保留小数点后2位|
	QuotaAmount float32 `json:"quota_amount"`
}

func (o AdjustCouponQuotasReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AdjustCouponQuotasReq struct{}"
	}

	return strings.Join([]string{"AdjustCouponQuotasReq", string(data)}, " ")
}
