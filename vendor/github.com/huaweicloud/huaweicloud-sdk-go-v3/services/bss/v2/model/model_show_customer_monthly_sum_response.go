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

// Response Object
type ShowCustomerMonthlySumResponse struct {
	// |参数名称：总条数，必须大于等于0。| |参数的约束及描述：总条数，必须大于等于0。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：账单记录，具体参考表 BillSumRecordInfo。| |参数约束以及描述：账单记录，具体参考表 BillSumRecordInfo。|
	BillSums *[]BillSumRecordInfoV2 `json:"bill_sums,omitempty"`
	// |参数名称：总金额（包含退订）。| |参数的约束及描述：总金额（包含退订）。|
	ConsumeAmount float32 `json:"consume_amount,omitempty"`
	// |参数名称：总欠费金额。| |参数的约束及描述：总欠费金额。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：代金券金额。| |参数的约束及描述：代金券金额。|
	CouponAmount float32 `json:"coupon_amount,omitempty"`
	// |参数名称：现金券金额，预留。| |参数的约束及描述：现金券金额，预留。|
	FlexipurchaseCouponAmount float32 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：储值卡金额，预留。| |参数的约束及描述：储值卡金额，预留。|
	StoredValueCardAmount float32 `json:"stored_value_card_amount,omitempty"`
	// |参数名称：现金账户金额。| |参数的约束及描述：现金账户金额。|
	CashAmount float32 `json:"cash_amount,omitempty"`
	// |参数名称：信用账户金额。| |参数的约束及描述：信用账户金额。|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：欠费核销金额| |参数的约束及描述：欠费核销金额|
	WriteoffAmount float32 `json:"writeoff_amount,omitempty"`
	// |参数名称：金额单位。1：元| |参数的约束及描述：金额单位。1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：币种。CNY：人民币。USD：美元。| |参数约束及描述：币种。CNY：人民币。USD：美元。|
	Currency       *string `json:"currency,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowCustomerMonthlySumResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerMonthlySumResponse struct{}"
	}

	return strings.Join([]string{"ShowCustomerMonthlySumResponse", string(data)}, " ")
}
