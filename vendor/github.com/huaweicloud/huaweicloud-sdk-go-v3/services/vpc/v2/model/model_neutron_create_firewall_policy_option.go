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
type NeutronCreateFirewallPolicyOption struct {
	// 审计标记。
	Audited *bool `json:"audited,omitempty"`
	// 功能说明：网络ACL防火墙策略描述 取值范围：最长255个字符
	Description *string `json:"description,omitempty"`
	// 策略引用的网络ACL防火墙规则链。
	FirewallRules *[]string `json:"firewall_rules,omitempty"`
	// 功能说明：网络ACL防火墙策略名称 取值范围：最长255个字符
	Name *string `json:"name,omitempty"`
}

func (o NeutronCreateFirewallPolicyOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallPolicyOption struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallPolicyOption", string(data)}, " ")
}
