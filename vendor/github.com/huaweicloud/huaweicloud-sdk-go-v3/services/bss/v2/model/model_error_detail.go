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

type ErrorDetail struct {
	// |参数名称：返回码| |参数的约束及描述：该参数非必填，最大长度16|
	ErrorCode *string `json:"error_code,omitempty"`
	// |参数名称：返回码描述| |参数的约束及描述：该参数非必填，最大长度1024|
	ErrorMsg *string `json:"error_msg,omitempty"`
	// |参数名称：标示ID| |参数的约束及描述：该参数非必填，最大长度256|
	Id *string `json:"id,omitempty"`
}

func (o ErrorDetail) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ErrorDetail struct{}"
	}

	return strings.Join([]string{"ErrorDetail", string(data)}, " ")
}
