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
type ListProvincesResponse struct {
	// |参数名称：查询个数，成功的时候返回| |参数的约束及描述：查询个数，成功的时候返回|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：省份信息列表，成功的时候返回| |参数约束以及描述：省份信息列表，成功的时候返回|
	Provinces      *[]Province `json:"provinces,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ListProvincesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProvincesResponse struct{}"
	}

	return strings.Join([]string{"ListProvincesResponse", string(data)}, " ")
}
