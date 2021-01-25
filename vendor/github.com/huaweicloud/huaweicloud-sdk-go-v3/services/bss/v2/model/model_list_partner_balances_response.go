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
type ListPartnerBalancesResponse struct {
	// |参数名称：总额，即最终优惠后的金额，| |参数约束以及描述：总额，即最终优惠后的金额，|
	AccountBalances *[]AccountBalanceV2 `json:"account_balances,omitempty"`
	HttpStatusCode  int                 `json:"-"`
}

func (o ListPartnerBalancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerBalancesResponse struct{}"
	}

	return strings.Join([]string{"ListPartnerBalancesResponse", string(data)}, " ")
}
