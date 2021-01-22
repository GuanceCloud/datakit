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

type UpdatePortRequestBody struct {
	// 新端口号。端口号有效范围为2100~9500。
	Port int32 `json:"port"`
}

func (o UpdatePortRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePortRequestBody struct{}"
	}

	return strings.Join([]string{"UpdatePortRequestBody", string(data)}, " ")
}
