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
type ListSubCustomersResponse struct {
	// |参数名称：客户信息列表。具体请参见表 CustomerInfo| |参数约束以及描述：客户信息列表。具体请参见表 CustomerInfo|
	CustomerInfos *[]CustomerInformation `json:"customer_infos,omitempty"`
	// |参数名称：总记录数。| |参数的约束及描述：总记录数。|
	Count          *int32 `json:"count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListSubCustomersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomersResponse struct{}"
	}

	return strings.Join([]string{"ListSubCustomersResponse", string(data)}, " ")
}
