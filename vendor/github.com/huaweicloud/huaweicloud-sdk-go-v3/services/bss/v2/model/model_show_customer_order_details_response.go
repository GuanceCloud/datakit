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
type ShowCustomerOrderDetailsResponse struct {
	// |参数名称：符合条件的记录总数。| |参数的约束及描述：符合条件的记录总数。|
	TotalCount *int32           `json:"total_count,omitempty"`
	OrderInfo  *CustomerOrderV2 `json:"order_info,omitempty"`
	// |参数名称：订单对应的订单项。具体请参见表 OrderLineItemEntity。| |参数约束及描述： 订单对应的订单项。具体请参见表 OrderLineItemEntity。|
	OrderLineItems *[]OrderLineItemEntityV2 `json:"order_line_items,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o ShowCustomerOrderDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerOrderDetailsResponse struct{}"
	}

	return strings.Join([]string{"ShowCustomerOrderDetailsResponse", string(data)}, " ")
}
