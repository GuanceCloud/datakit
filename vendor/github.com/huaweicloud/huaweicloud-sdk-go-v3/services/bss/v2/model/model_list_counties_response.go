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
type ListCountiesResponse struct {
	// |参数名称：查询个数，成功的时候返回| |参数的约束及描述：查询个数，成功的时候返回|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：区县信息列表，成功的时候返回| |参数约束以及描述：区县信息列表，成功的时候返回|
	Counties       *[]County `json:"counties,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListCountiesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCountiesResponse struct{}"
	}

	return strings.Join([]string{"ListCountiesResponse", string(data)}, " ")
}
