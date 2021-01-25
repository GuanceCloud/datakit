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

type RetrieveAmountInfoV2 struct {
	// |参数名称：可回收的金额。| |参数的约束及描述：可回收的金额。|
	AvailRetrieveAmount float32 `json:"avail_retrieve_amount,omitempty"`
	// |参数名称：金额单位。1：元| |参数的约束及描述：金额单位。1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：币种。CNY：人民币USD：美金| |参数约束及描述：币种。CNY：人民币USD：美金|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：账户余额（仅balance_type=信用账户时才有这个字段）。| |参数的约束及描述：账户余额（仅balance_type=信用账户时才有这个字段）。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：信用额度（仅balance_type=信用账户时才有这个字段）。| |参数的约束及描述：信用额度（仅balance_type=信用账户时才有这个字段）。|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：信用额度过期时间。UTC时间，格式为：2016-03-28T14:45:38Z。如果查询信用账户可回收余额的查询结果没有失效时间，表示永久有效。| |参数约束及描述：信用额度过期时间。UTC时间，格式为：2016-03-28T14:45:38Z。如果查询信用账户可回收余额的查询结果没有失效时间，表示永久有效。|
	ExpireTime *string `json:"expire_time,omitempty"`
}

func (o RetrieveAmountInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RetrieveAmountInfoV2 struct{}"
	}

	return strings.Join([]string{"RetrieveAmountInfoV2", string(data)}, " ")
}
