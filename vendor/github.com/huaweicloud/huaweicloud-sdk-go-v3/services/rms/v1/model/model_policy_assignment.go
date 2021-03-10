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

// 规则
type PolicyAssignment struct {
	// 规则ID
	Id *string `json:"id,omitempty"`
	// 规则名字
	Name *string `json:"name,omitempty"`
	// 规则描述
	Description  *string                 `json:"description,omitempty"`
	PolicyFilter *PolicyFilterDefinition `json:"policy_filter,omitempty"`
	// 规则状态
	State *string `json:"state,omitempty"`
	// 规则创建时间
	Created *string `json:"created,omitempty"`
	// 规则更新时间
	Updated *string `json:"updated,omitempty"`
	// 规则的策略ID
	PolicyDefinitionId *string `json:"policy_definition_id,omitempty"`
	// 规则参数
	Parameters map[string]PolicyParameterValue `json:"parameters,omitempty"`
}

func (o PolicyAssignment) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyAssignment struct{}"
	}

	return strings.Join([]string{"PolicyAssignment", string(data)}, " ")
}
