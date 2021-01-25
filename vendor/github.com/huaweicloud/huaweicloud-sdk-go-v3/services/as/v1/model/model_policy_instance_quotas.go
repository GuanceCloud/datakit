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

// 配额列表
type PolicyInstanceQuotas struct {
	// 配额资源详情。
	Resources *[]PolicyInstanceResources `json:"resources,omitempty"`
}

func (o PolicyInstanceQuotas) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyInstanceQuotas struct{}"
	}

	return strings.Join([]string{"PolicyInstanceQuotas", string(data)}, " ")
}
