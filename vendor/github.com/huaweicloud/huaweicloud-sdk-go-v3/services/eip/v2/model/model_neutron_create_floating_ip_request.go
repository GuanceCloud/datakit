/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type NeutronCreateFloatingIpRequest struct {
	Body *NeutronCreateFloatingIpRequestBody `json:"body,omitempty"`
}

func (o NeutronCreateFloatingIpRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFloatingIpRequest struct{}"
	}

	return strings.Join([]string{"NeutronCreateFloatingIpRequest", string(data)}, " ")
}
