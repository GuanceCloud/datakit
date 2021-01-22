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

type City struct {
	// |参数名称：城市的编码。| |参数约束及描述：城市的编码。|
	Code string `json:"code"`
	// |参数名称：城市的名称，根据请求的语言会传递回对应的语言的名称，目前仅支持中文。| |参数约束及描述：城市的名称，根据请求的语言会传递回对应的语言的名称，目前仅支持中文。|
	Name string `json:"name"`
}

func (o City) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "City struct{}"
	}

	return strings.Join([]string{"City", string(data)}, " ")
}
