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

type DataIpRequest struct {
	// 内网ip
	NewIp string `json:"new_ip"`
}

func (o DataIpRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DataIpRequest struct{}"
	}

	return strings.Join([]string{"DataIpRequest", string(data)}, " ")
}
