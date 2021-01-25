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

// Request Object
type ListOrderCouponsByOrderIdRequest struct {
	OrderId string `json:"order_id"`
}

func (o ListOrderCouponsByOrderIdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOrderCouponsByOrderIdRequest struct{}"
	}

	return strings.Join([]string{"ListOrderCouponsByOrderIdRequest", string(data)}, " ")
}
