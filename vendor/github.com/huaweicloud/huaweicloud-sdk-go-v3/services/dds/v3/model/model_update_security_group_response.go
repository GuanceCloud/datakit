/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateSecurityGroupResponse struct {
	// 任务ID。
	JobId *string `json:"job_id,omitempty"`
	// 实例当前安全组。
	SecurityGroupId *string `json:"security_group_id,omitempty"`
	HttpStatusCode  int     `json:"-"`
}

func (o UpdateSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"UpdateSecurityGroupResponse", string(data)}, " ")
}
