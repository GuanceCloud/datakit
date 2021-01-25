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
type ListPayPerUseCustomerResourcesResponse struct {
	// |参数名称：资源列表。具体请参见表2 OrderInstanceV2。| |参数约束以及描述：资源列表。具体请参见表2 OrderInstanceV2。|
	Data *[]OrderInstanceV2 `json:"data,omitempty"`
	// |参数名称：总记录数。| |参数的约束及描述：总记录数。|
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListPayPerUseCustomerResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPayPerUseCustomerResourcesResponse struct{}"
	}

	return strings.Join([]string{"ListPayPerUseCustomerResourcesResponse", string(data)}, " ")
}
