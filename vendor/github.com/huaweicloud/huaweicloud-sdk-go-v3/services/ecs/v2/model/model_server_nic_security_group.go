/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type ServerNicSecurityGroup struct {
	// 安全组ID。
	Id string `json:"id"`
}

func (o ServerNicSecurityGroup) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServerNicSecurityGroup struct{}"
	}

	return strings.Join([]string{"ServerNicSecurityGroup", string(data)}, " ")
}
