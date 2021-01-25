/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SecurityGroupRequest struct {
	// - 安全组ID。
	SecurityGroupId string `json:"security_group_id"`
}

func (o SecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"SecurityGroupRequest", string(data)}, " ")
}
