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
type BatchCreateSubNetworkInterfaceOption struct {
	// 功能说明：虚拟子网ID 取值范围：标准UUID
	VirsubnetId string `json:"virsubnet_id"`
	// 功能说明：宿主网络接口的ID 取值范围：标注UUID 约束：必须是实际存在的端口ID
	ParentId string `json:"parent_id"`
	// 功能说明：安全组的ID列表；例如：\"security_groups\": [\"a0608cbf-d047-4f54-8b28-cd7b59853fff\"] 取值范围：默认值为系统默认安全组
	SecurityGroups *[]string `json:"security_groups,omitempty"`
	// 功能说明：辅助弹性网卡的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description *string `json:"description,omitempty"`
	// 功能说明：辅助弹性网卡是否启用ipv6地址 取值范围：true（开启)，false（关闭） 默认值：false
	Ipv6Enable *bool `json:"ipv6_enable,omitempty"`
	// 功能说明：辅助弹性网卡所属的项目ID 取值范围：标准UUID 约束：只有管理员有权限指定
	ProjectId *string `json:"project_id,omitempty"`
}

func (o BatchCreateSubNetworkInterfaceOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateSubNetworkInterfaceOption struct{}"
	}

	return strings.Join([]string{"BatchCreateSubNetworkInterfaceOption", string(data)}, " ")
}
