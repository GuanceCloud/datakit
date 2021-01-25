/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 创建floatingip对象
type NeutronCreateFloatingIpRequestBody struct {
	Floatingip *CreateFloatingIpOption `json:"floatingip"`
}

func (o NeutronCreateFloatingIpRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFloatingIpRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronCreateFloatingIpRequestBody", string(data)}, " ")
}
