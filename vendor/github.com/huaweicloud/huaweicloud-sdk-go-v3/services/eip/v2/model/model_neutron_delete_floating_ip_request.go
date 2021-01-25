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
type NeutronDeleteFloatingIpRequest struct {
	FloatingipId string `json:"floatingip_id"`
}

func (o NeutronDeleteFloatingIpRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFloatingIpRequest struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFloatingIpRequest", string(data)}, " ")
}
