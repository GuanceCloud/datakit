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
type ListCitiesResponse struct {
	// |参数名称：查询个数，成功的时候返回| |参数的约束及描述：查询个数，成功的时候返回|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：城市信息列表，成功的时候返回| |参数约束以及描述：城市信息列表，成功的时候返回|
	Cities         *[]City `json:"cities,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListCitiesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCitiesResponse struct{}"
	}

	return strings.Join([]string{"ListCitiesResponse", string(data)}, " ")
}
