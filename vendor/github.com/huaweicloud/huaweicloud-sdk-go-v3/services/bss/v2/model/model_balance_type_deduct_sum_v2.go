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

type BalanceTypeDeductSumV2 struct {
	// |参数名称：账户类型。BALANCE_TYPE_DEBIT：余额BALANCE_TYPE_CREDIT：信用BALANCE_TYPE_BONUS：奖励BALANCE_TYPE_COUPON：代金券BALANCE_TYPE_RCASH_COUPON 现金券。BALANCE_TYPE_STORED_VALUE_CARD：储值卡消费| |参数约束及描述：账户类型。BALANCE_TYPE_DEBIT：余额BALANCE_TYPE_CREDIT：信用BALANCE_TYPE_BONUS：奖励BALANCE_TYPE_COUPON：代金券BALANCE_TYPE_RCASH_COUPON 现金券。BALANCE_TYPE_STORED_VALUE_CARD：储值卡消费|
	BalanceType *string `json:"balance_type,omitempty"`
	// |参数名称：金额。对于billType=1或者2的账单，该金额为负值。| |参数的约束及描述：金额。对于billType=1或者2的账单，该金额为负值。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：费用类型。0：正常；1：退订；2：华为核销。| |参数约束及描述：费用类型。0：正常；1：退订；2：华为核销。|
	BillType *string `json:"bill_type,omitempty"`
}

func (o BalanceTypeDeductSumV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BalanceTypeDeductSumV2 struct{}"
	}

	return strings.Join([]string{"BalanceTypeDeductSumV2", string(data)}, " ")
}
