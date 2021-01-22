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

// 更新floatingip的请求体
type NeutronUpdateFloatingIpRequestBody struct {
	Floatingip *UpdateFloatingIpOption `json:"floatingip"`
}

func (o NeutronUpdateFloatingIpRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFloatingIpRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFloatingIpRequestBody", string(data)}, " ")
}
