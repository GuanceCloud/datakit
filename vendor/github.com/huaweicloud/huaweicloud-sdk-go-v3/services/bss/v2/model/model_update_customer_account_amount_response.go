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
type UpdateCustomerAccountAmountResponse struct {
	// |参数名称：总额，即最终优惠后的金额，| |参数约束及描述：总额，即最终优惠后的金额，|
	TransferId     *string `json:"transfer_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateCustomerAccountAmountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateCustomerAccountAmountResponse struct{}"
	}

	return strings.Join([]string{"UpdateCustomerAccountAmountResponse", string(data)}, " ")
}
