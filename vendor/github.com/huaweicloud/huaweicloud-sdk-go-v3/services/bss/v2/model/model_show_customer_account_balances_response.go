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
type ShowCustomerAccountBalancesResponse struct {
	// |参数名称：账户余额列表。具体请参见表 AccountBalanceV3| |参数约束以及描述：账户余额列表。具体请参见表 AccountBalanceV3|
	AccountBalances *[]AccountBalanceV3 `json:"account_balances,omitempty"`
	// |参数名称：欠款总金额。| |参数的约束及描述：欠款总金额。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：度量单位：1：元2：角3：分| |参数的约束及描述：度量单位：1：元2：角3：分|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：币种。CNY：人民币。USD：美元。| |参数约束及描述：币种。CNY：人民币。USD：美元。|
	Currency       *string `json:"currency,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowCustomerAccountBalancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerAccountBalancesResponse struct{}"
	}

	return strings.Join([]string{"ShowCustomerAccountBalancesResponse", string(data)}, " ")
}
