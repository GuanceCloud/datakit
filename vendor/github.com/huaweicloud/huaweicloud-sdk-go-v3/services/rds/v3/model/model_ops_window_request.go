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

type OpsWindowRequest struct {
	// - 开始时间， UTC时间
	StartTime string `json:"start_time"`
	// - 结束时间，UTC时间
	EndTime string `json:"end_time"`
}

func (o OpsWindowRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OpsWindowRequest struct{}"
	}

	return strings.Join([]string{"OpsWindowRequest", string(data)}, " ")
}
