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
type NovaServerSecurityGroup struct {
	// 安全组名称或者uuid。
	Name *string `json:"name,omitempty"`
}

func (o NovaServerSecurityGroup) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaServerSecurityGroup struct{}"
	}

	return strings.Join([]string{"NovaServerSecurityGroup", string(data)}, " ")
}
