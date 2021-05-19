/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 策略定义
type PolicyDefinition struct {
	// 策略id
	Id *string `json:"id,omitempty"`
	// 策略名字
	Name *string `json:"name,omitempty"`
	// 策略类型
	PolicyType *string `json:"policy_type,omitempty"`
	// 策略描述
	Description *string `json:"description,omitempty"`
	// 策略语法类型
	PolicyRuleType *string `json:"policy_rule_type,omitempty"`
	// 策略规则
	PolicyRule *interface{} `json:"policy_rule,omitempty"`
	// 关键词列表
	Keywords *[]string `json:"keywords,omitempty"`
	// 策略参数
	Parameters map[string]PolicyParameterDefinition `json:"parameters,omitempty"`
}

func (o PolicyDefinition) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyDefinition struct{}"
	}

	return strings.Join([]string{"PolicyDefinition", string(data)}, " ")
}
