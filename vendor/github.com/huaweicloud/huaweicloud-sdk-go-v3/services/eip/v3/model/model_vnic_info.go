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

// 公网IP中的vnic对象，存储绑定PORT的相关信息
type VnicInfo struct {
	// 功能说明：私网IP地址
	PrivateIpAddress *string `json:"private_ip_address,omitempty"`
	// 功能说明：端口所属设备ID 约束：不支持设置和更新,由系统自动维护
	DeviceId *string `json:"device_id,omitempty"`
	// 功能说明：设备所属 取值范围：合法设备所属,如network:dhcp、network:VIP_PORT、network:router_interface_distributed、network:router_centralized_snat 约束：不支持设置和更新,由系统自动维护
	DeviceOwner *string `json:"device_owner,omitempty"`
	// 功能说明：虚拟私有云ID
	VpcId *string `json:"vpc_id,omitempty"`
	// 功能说明：端口ID
	PortId *string `json:"port_id,omitempty"`
	// 功能说明：端口profile信息
	PortProfile *string `json:"port_profile,omitempty"`
	// 功能说明：端口MAC地址 约束：由系统分配,不支持指定
	Mac *string `json:"mac,omitempty"`
	// 功能说明：VTEP IP
	Vtep *string `json:"vtep,omitempty"`
	// 功能说明：VXLAN ID
	Vni *string `json:"vni,omitempty"`
	// 功能说明：端口所属实例ID,例如RDS实例ID 约束：不支持设置和更新,由系统自动维护
	InstanceId *string `json:"instance_id,omitempty"`
	// 功能说明：端口所属实例类型,例如“RDS” 约束：不支持设置和更新,由系统自动维护
	InstanceType *string `json:"instance_type,omitempty"`
}

func (o VnicInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VnicInfo struct{}"
	}

	return strings.Join([]string{"VnicInfo", string(data)}, " ")
}
