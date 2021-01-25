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

type CustomerBalancesV2 struct {
	// |参数名称：客户的客户ID。| |参数约束及描述：客户的客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：客户欠款总额度。| |参数约束及描述： 客户欠款总额度。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：客户可用总额度。| |参数约束及描述： 客户可用总额度。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：币种。| |参数约束及描述：币种。|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：度量单位：1：元；2：角；3：分。| |参数的约束及描述：度量单位：1：元；2：角；3：分。|
	MeasureId *int32 `json:"measure_id,omitempty"`
}

func (o CustomerBalancesV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CustomerBalancesV2 struct{}"
	}

	return strings.Join([]string{"CustomerBalancesV2", string(data)}, " ")
}
