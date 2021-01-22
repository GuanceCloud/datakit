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
type ListCustomersBalancesDetailResponse struct {
	// |参数名称：总额，即最终优惠后的金额，| |参数约束以及描述：总额，即最终优惠后的金额，|
	CustomerBalances *[]CustomerBalancesV2 `json:"customer_balances,omitempty"`
	HttpStatusCode   int                   `json:"-"`
}

func (o ListCustomersBalancesDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomersBalancesDetailResponse struct{}"
	}

	return strings.Join([]string{"ListCustomersBalancesDetailResponse", string(data)}, " ")
}
