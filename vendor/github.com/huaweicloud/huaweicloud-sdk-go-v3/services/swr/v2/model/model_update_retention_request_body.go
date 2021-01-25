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

type UpdateRetentionRequestBody struct {
	// 算法
	Algorithm string `json:"algorithm"`
	// 镜像老化规则
	Rules []Rule `json:"rules"`
}

func (o UpdateRetentionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateRetentionRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateRetentionRequestBody", string(data)}, " ")
}
