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

// Response Object
type ListPolicyAssignmentsResponse struct {
	// 规则列表
	Value          *[]PolicyAssignment `json:"value,omitempty"`
	PageInfo       *PageInfo           `json:"page_info,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListPolicyAssignmentsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPolicyAssignmentsResponse struct{}"
	}

	return strings.Join([]string{"ListPolicyAssignmentsResponse", string(data)}, " ")
}
