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

type LimitValue struct {
	// |参数名称：属性值1| |参数约束及描述：属性值1|
	Value1 *string `json:"value1,omitempty"`
	// |参数名称：属性值2| |参数约束及描述：属性值2|
	Value2 *string `json:"value2,omitempty"`
}

func (o LimitValue) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "LimitValue struct{}"
	}

	return strings.Join([]string{"LimitValue", string(data)}, " ")
}
