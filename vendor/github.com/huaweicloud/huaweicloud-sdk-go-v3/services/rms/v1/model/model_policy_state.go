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

// 合规状态
type PolicyState struct {
	// 合规状态所属用户ID
	DomainId *string `json:"domain_id,omitempty"`
	// 合规状态所属资源区域ID
	RegionId *string `json:"region_id,omitempty"`
	// 合规状态所属资源ID
	ResourceId *string `json:"resource_id,omitempty"`
	// 合规状态所属资源名字
	ResourceName *string `json:"resource_name,omitempty"`
	// 合规状态所属资源provider
	ResourceProvider *string `json:"resource_provider,omitempty"`
	// 合规状态所属资源类型
	ResourceType *string `json:"resource_type,omitempty"`
	// 合规状态
	ComplianceState *string `json:"compliance_state,omitempty"`
	// 合规状态所属规则ID
	PolicyAssignmentId *string `json:"policy_assignment_id,omitempty"`
	// 合规状态所属规则名字
	PolicyAssignmentName *string `json:"policy_assignment_name,omitempty"`
	// 合规状态所属策略ID
	PolicyDefinitionId *string `json:"policy_definition_id,omitempty"`
	// 合规状态评估时间
	EvaluationTime *string `json:"evaluation_time,omitempty"`
}

func (o PolicyState) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyState struct{}"
	}

	return strings.Join([]string{"PolicyState", string(data)}, " ")
}
