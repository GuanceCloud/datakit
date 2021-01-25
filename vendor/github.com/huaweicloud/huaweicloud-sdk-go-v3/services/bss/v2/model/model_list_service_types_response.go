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
type ListServiceTypesResponse struct {
	// |参数名称：返回数据| |参数约束以及描述：返回数据|
	ServiceTypes   *[]ServiceType `json:"service_types,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListServiceTypesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServiceTypesResponse struct{}"
	}

	return strings.Join([]string{"ListServiceTypesResponse", string(data)}, " ")
}
