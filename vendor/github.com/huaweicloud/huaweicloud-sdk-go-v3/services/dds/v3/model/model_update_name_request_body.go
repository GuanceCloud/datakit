/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type UpdateNameRequestBody struct {
	// 新实例名称。用于表示实例的名称，同一租户下，同类型的实例名唯一。取值范围：长度为4~64位，必须以字母开头（A~Z或a~z），区分大小写，可以包含字母、数字（0~9）、中划线（-）或者下划线（_），不能包含其他特殊字符。
	NewInstanceName string `json:"new_instance_name"`
}

func (o UpdateNameRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNameRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateNameRequestBody", string(data)}, " ")
}
