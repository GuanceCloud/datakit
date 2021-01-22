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
type ListEnterpriseSubCustomersResponse struct {
	// |参数名称：结果集数量，成功才有。| |参数的约束及描述：结果集数量，成功才有。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：客户信息列表，成功才有。| |参数约束以及描述：客户信息列表，成功才有。|
	SubCustomerInfos *[]SubCustomerInfoV2 `json:"sub_customer_infos,omitempty"`
	HttpStatusCode   int                  `json:"-"`
}

func (o ListEnterpriseSubCustomersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnterpriseSubCustomersResponse struct{}"
	}

	return strings.Join([]string{"ListEnterpriseSubCustomersResponse", string(data)}, " ")
}
