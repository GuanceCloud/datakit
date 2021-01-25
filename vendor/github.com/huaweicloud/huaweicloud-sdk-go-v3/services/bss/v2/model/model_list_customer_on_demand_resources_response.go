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
type ListCustomerOnDemandResourcesResponse struct {
	// |参数名称：客户按需资源列表。CustomerOnDemandResource| |参数约束以及描述：客户按需资源列表。CustomerOnDemandResource|
	Resources *[]CustomerOnDemandResource `json:"resources,omitempty"`
	// |参数名称：查询总数| |参数的约束及描述：查询总数|
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListCustomerOnDemandResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerOnDemandResourcesResponse struct{}"
	}

	return strings.Join([]string{"ListCustomerOnDemandResourcesResponse", string(data)}, " ")
}
