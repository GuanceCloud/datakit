/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 安全组信息
type SecurityGroups struct {
	// 安全组ID。
	Id string `json:"id"`
}

func (o SecurityGroups) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroups struct{}"
	}

	return strings.Join([]string{"SecurityGroups", string(data)}, " ")
}
