/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type UpdateSubNetworkInterfaceOption struct {
	// 功能说明：辅助弹性网卡的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description *string `json:"description,omitempty"`
	// 功能说明：安全组的ID列表；例如：\"security_groups\": [\"a0608cbf-d047-4f54-8b28-cd7b59853fff\"]
	SecurityGroups *[]string `json:"security_groups,omitempty"`
}

func (o UpdateSubNetworkInterfaceOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubNetworkInterfaceOption struct{}"
	}

	return strings.Join([]string{"UpdateSubNetworkInterfaceOption", string(data)}, " ")
}
