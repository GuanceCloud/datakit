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

// 规则请求体
type PolicyAssignmentRequestBody struct {
	// 规则名字
	Name *string `json:"name,omitempty"`
	// 规则描述
	Description  *string                 `json:"description,omitempty"`
	PolicyFilter *PolicyFilterDefinition `json:"policy_filter,omitempty"`
	// 策略定义ID
	PolicyDefinitionId *string `json:"policy_definition_id,omitempty"`
	// 规则参数
	Parameters map[string]PolicyParameterValue `json:"parameters,omitempty"`
}

func (o PolicyAssignmentRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyAssignmentRequestBody struct{}"
	}

	return strings.Join([]string{"PolicyAssignmentRequestBody", string(data)}, " ")
}
