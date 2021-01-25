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

type ApplyConfigurationRequest struct {
	// 实例ID列表。
	InstanceIds []string `json:"instance_ids"`
}

func (o ApplyConfigurationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplyConfigurationRequest struct{}"
	}

	return strings.Join([]string{"ApplyConfigurationRequest", string(data)}, " ")
}
