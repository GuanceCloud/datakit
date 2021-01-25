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

type ReclaimCouponQuotasReq struct {
	// |参数名称：被回收的代金券额度的ID。| |参数约束以及描述：被回收的代金券额度的ID。|
	QuotaIds []string `json:"quota_ids"`
	// |参数名称：回收时候的备注| |参数约束及描述：回收时候的备注|
	Remark *string `json:"remark,omitempty"`
}

func (o ReclaimCouponQuotasReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimCouponQuotasReq struct{}"
	}

	return strings.Join([]string{"ReclaimCouponQuotasReq", string(data)}, " ")
}
