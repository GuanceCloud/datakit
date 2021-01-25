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
type BatchSetSubCustomerDiscountResponse struct {
	// |参数名称：错误的客户列表和错误信息| |参数约束以及描述：错误的客户列表和错误信息|
	ErrorDetails   *[]ErrorDetail `json:"error_details,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o BatchSetSubCustomerDiscountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchSetSubCustomerDiscountResponse struct{}"
	}

	return strings.Join([]string{"BatchSetSubCustomerDiscountResponse", string(data)}, " ")
}
