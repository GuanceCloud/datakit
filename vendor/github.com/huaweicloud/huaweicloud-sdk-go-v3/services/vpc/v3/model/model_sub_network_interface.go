/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

//
type SubNetworkInterface struct {
	// 功能说明：辅助弹性网卡的唯一标识 取值范围：带(-)的标准UUID
	Id string `json:"id"`
	// 功能说明：虚拟子网ID 取值范围：标准UUID
	VirsubnetId string `json:"virsubnet_id"`
	// 功能说明：辅助弹性网卡的私有IPv4地址 取值范围：必须在虚拟子网的网段内，不填则随机在虚拟子网网段内随机分配
	PrivateIpAddress string `json:"private_ip_address"`
	// 功能说明：辅助弹性网卡的IPv6地址
	Ipv6IpAddress string `json:"ipv6_ip_address"`
	// 功能说明：辅助弹性网卡的mac地址 取值范围：合法的mac地址，系统随机分配
	MacAddress string `json:"mac_address"`
	// 功能说明：设备ID 取值范围：标准UUID
	ParentDeviceId string `json:"parent_device_id"`
	// 功能说明：宿主网络接口的ID 取值范围：标准UUID
	ParentId string `json:"parent_id"`
	// 功能说明：辅助弹性网卡的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description string `json:"description"`
	// 功能说明：辅助弹性网卡所属的VPC_ID 取值范围：标准UUID
	VpcId string `json:"vpc_id"`
	// 功能说明：辅助弹性网卡的VLAN ID 取值范围：1-4094 约束：同一个宿主网络接口下唯一
	VlanId int32 `json:"vlan_id"`
	// 功能说明：安全组的ID列表；例如：\"security_groups\": [\"a0608cbf-d047-4f54-8b28-cd7b59853fff\"] 取值范围：默认值为系统默认安全组
	SecurityGroups []string `json:"security_groups"`
	// 功能说明：辅助弹性网卡的标签列表
	Tags []string `json:"tags"`
	// 功能说明：辅助弹性网卡所属项目ID
	ProjectId string `json:"project_id"`
	// 功能说明：辅助弹性网卡的创建时间 取值范围：UTC时间格式：yyyy-MM-ddTHH:mm:ss
	CreatedAt *sdktime.SdkTime `json:"created_at"`
}

func (o SubNetworkInterface) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SubNetworkInterface struct{}"
	}

	return strings.Join([]string{"SubNetworkInterface", string(data)}, " ")
}
