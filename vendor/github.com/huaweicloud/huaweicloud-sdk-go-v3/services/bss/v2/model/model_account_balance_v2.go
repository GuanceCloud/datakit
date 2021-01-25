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

type AccountBalanceV2 struct {
	// |参数名称：账户标识。| |参数约束及描述：账户标识。|
	AccountId string `json:"account_id"`
	// |参数名称：账户类型：1：余额2：信用5：奖励7：保证金8：可拨款| |参数的约束及描述：账户类型：1：余额2：信用5：奖励7：保证金8：可拨款|
	AccountType int32 `json:"account_type"`
	// |参数名称：余额。| |参数的约束及描述：余额。|
	Amount float64 `json:"amount"`
	// |参数名称：币种。当前固定为CNY。| |参数约束及描述：币种。当前固定为CNY。|
	Currency string `json:"currency"`
	// |参数名称：专款专用余额。| |参数的约束及描述：专款专用余额。|
	DesignatedAmount *float64 `json:"designated_amount,omitempty"`
	// |参数名称：总信用额度。只有账户类型是2:信用的时候才有该字段| |参数的约束及描述：总信用额度。只有账户类型是2:信用的时候才有该字段|
	CreditAmount *float64 `json:"credit_amount,omitempty"`
	// |参数名称：度量单位。1：元。| |参数的约束及描述：度量单位。1：元。|
	MeasureId int32 `json:"measure_id"`
	// |参数名称：备注。| |参数约束及描述：备注。|
	Memo *string `json:"memo,omitempty"`
}

func (o AccountBalanceV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AccountBalanceV2 struct{}"
	}

	return strings.Join([]string{"AccountBalanceV2", string(data)}, " ")
}
