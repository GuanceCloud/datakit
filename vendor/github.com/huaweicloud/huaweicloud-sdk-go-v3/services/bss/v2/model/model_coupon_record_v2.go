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

type CouponRecordV2 struct {
	// |参数名称：唯一ID。| |参数约束及描述：唯一ID。|
	Id *string `json:"id,omitempty"`
	// |参数名称：操作类型：1：发放；2：手动回收；3：解绑自动回收| |参数约束及描述：操作类型：1：发放；2：手动回收；3：解绑自动回收|
	OperationType *string `json:"operation_type,omitempty"`
	// |参数名称：额度ID。| |参数约束及描述：额度ID。|
	QuotaId *string `json:"quota_id,omitempty"`
	// |参数名称：操作类型：1：发放；2：手动回收；3：解绑自动回收。| |参数的约束及描述：操作类型：1：发放；2：手动回收；3：解绑自动回收。|
	QuotaType *int32 `json:"quota_type,omitempty"`
	// |参数名称：代金券ID。| |参数约束及描述：代金券ID。|
	CouponId *string `json:"coupon_id,omitempty"`
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：操作的面额值。发放时，等于面额值；回收时，指每次回收的具体值。| |参数的约束及描述：操作的面额值。发放时，等于面额值；回收时，指每次回收的具体值。|
	OperationAmount float32 `json:"operation_amount,omitempty"`
	// |参数名称：操作时间。| |参数约束及描述：操作时间。|
	OperationTime *string `json:"operation_time,omitempty"`
	// |参数名称：操作结果：0：成功；其他：失败（直接记录错误码）。|参数约束及描述：操作结果：0：成功；其他：失败（直接记录错误码）。|
	Result *string `json:"result,omitempty"`
	// |参数名称：操作记录中的备注| |参数约束及描述：操作记录中的备注|
	Remark *string `json:"remark,omitempty"`
}

func (o CouponRecordV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CouponRecordV2 struct{}"
	}

	return strings.Join([]string{"CouponRecordV2", string(data)}, " ")
}
