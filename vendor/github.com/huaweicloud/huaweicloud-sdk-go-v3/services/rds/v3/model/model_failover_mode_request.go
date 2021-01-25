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

type FailoverModeRequest struct {
	// 数据库主备同步模式
	Mode string `json:"mode"`
}

func (o FailoverModeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FailoverModeRequest struct{}"
	}

	return strings.Join([]string{"FailoverModeRequest", string(data)}, " ")
}
