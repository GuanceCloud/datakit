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
type ListSubCustomerDiscountsResponse struct {
	SubCustomerDiscount *QuerySubCustomerDiscountV2 `json:"sub_customer_discount,omitempty"`
	HttpStatusCode      int                         `json:"-"`
}

func (o ListSubCustomerDiscountsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomerDiscountsResponse struct{}"
	}

	return strings.Join([]string{"ListSubCustomerDiscountsResponse", string(data)}, " ")
}
