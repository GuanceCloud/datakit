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

// 项目标签信息。
type ProjectTagInfoResponse struct {
	// 标签键。
	Key string `json:"key"`
	// 标签值列表。
	Values []string `json:"values"`
}

func (o ProjectTagInfoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ProjectTagInfoResponse struct{}"
	}

	return strings.Join([]string{"ProjectTagInfoResponse", string(data)}, " ")
}
