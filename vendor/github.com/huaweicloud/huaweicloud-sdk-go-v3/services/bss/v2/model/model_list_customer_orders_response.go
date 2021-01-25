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
type ListCustomerOrdersResponse struct {
	// |参数名称：符合条件的记录总数。| |参数的约束及描述：符合条件的记录总数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：客户订单详情信息。具体请参见表 CustomerOrderV2| |参数约束以及描述：客户订单详情信息。具体请参见表 CustomerOrderV2|
	OrderInfos     *[]CustomerOrderV2 `json:"order_infos,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListCustomerOrdersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerOrdersResponse struct{}"
	}

	return strings.Join([]string{"ListCustomerOrdersResponse", string(data)}, " ")
}
