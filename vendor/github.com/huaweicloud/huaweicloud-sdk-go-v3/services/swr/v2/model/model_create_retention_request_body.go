/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateRetentionRequestBody struct {
	// 回收规则匹配策略，or
	Algorithm string `json:"algorithm"`
	// 镜像老化规则
	Rules []Rule `json:"rules"`
}

func (o CreateRetentionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRetentionRequestBody struct{}"
	}

	return strings.Join([]string{"CreateRetentionRequestBody", string(data)}, " ")
}
