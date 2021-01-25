/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// absolute数据结构说明
type Absolute struct {
	// 裸金属服务器最大申请数量
	MaxTotalInstances *int32 `json:"maxTotalInstances,omitempty"`
	// CPU核数最大申请数量
	MaxTotalCores *int32 `json:"maxTotalCores,omitempty"`
	// 内存最大申请容量（单位：MB）
	MaxTotalRAMSize *int32 `json:"maxTotalRAMSize,omitempty"`
	// 可以申请的SSH密钥对最大数量
	MaxTotalKeypairs *int32 `json:"maxTotalKeypairs,omitempty"`
	// 可输入元数据的最大长度
	MaxServerMeta *int32 `json:"maxServerMeta,omitempty"`
	// 可注入文件的最大个数
	MaxPersonality *int32 `json:"maxPersonality,omitempty"`
	// 注入文件内容的最大长度（单位：Byte）
	MaxPersonalitySize *int32 `json:"maxPersonalitySize,omitempty"`
	// 服务器组的最大个数
	MaxServerGroups *int32 `json:"maxServerGroups,omitempty"`
	// 服务器组中的最大裸金属服务器数。
	MaxServerGroupMembers *int32 `json:"maxServerGroupMembers,omitempty"`
	// 已使用的服务器组个数
	TotalServerGroupsUsed *int32 `json:"totalServerGroupsUsed,omitempty"`
	// 安全组最大使用个数。 说明：具体配额限制请以VPC配额限制为准。
	MaxSecurityGroups *int32 `json:"maxSecurityGroups,omitempty"`
	// 安全组中安全组规则最大的配置个数。 说明：具体配额限制请以VPC配额限制为准。
	MaxSecurityGroupRules *int32 `json:"maxSecurityGroupRules,omitempty"`
	// 最大的浮动IP使用个数
	MaxTotalFloatingIps *int32 `json:"maxTotalFloatingIps,omitempty"`
	// 镜像元数据最大的长度
	MaxImageMeta *int32 `json:"maxImageMeta,omitempty"`
	// 当前裸金属服务器使用个数
	TotalInstancesUsed *int32 `json:"totalInstancesUsed,omitempty"`
	// 当前已使用CPU核数
	TotalCoresUsed *int32 `json:"totalCoresUsed,omitempty"`
	// 当前内存使用容量（单位：MB）
	TotalRAMUsed *int32 `json:"totalRAMUsed,omitempty"`
	// 当前安全组使用个数
	TotalSecurityGroupsUsed *int32 `json:"totalSecurityGroupsUsed,omitempty"`
	// 当前浮动IP使用个数
	TotalFloatingIpsUsed *int32 `json:"totalFloatingIpsUsed,omitempty"`
}

func (o Absolute) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Absolute struct{}"
	}

	return strings.Join([]string{"Absolute", string(data)}, " ")
}
