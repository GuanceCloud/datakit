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
type CreateFloatingIpOption struct {
	// 浮动IP地址。
	FloatingIpAddress *string `json:"floating_ip_address,omitempty"`
	// 外部网络的id。只能使用固定的外网，外部网络的信息请通过GET /v2.0/networks?router:external=True或GET /v2.0/networks?name={floating_network}或neutron net-external-list方式查询。
	FloatingNetworkId string `json:"floating_network_id"`
	// 端口id
	PortId *string `json:"port_id,omitempty"`
	// 关联端口的私有IP地址。
	FixedIpAddress *string `json:"fixed_ip_address,omitempty"`
}

func (o CreateFloatingIpOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateFloatingIpOption struct{}"
	}

	return strings.Join([]string{"CreateFloatingIpOption", string(data)}, " ")
}
