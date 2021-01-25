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

// Response Object
type NeutronShowFloatingIpResponse struct {
	Floatingip     *FloatingIpResp `json:"floatingip,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o NeutronShowFloatingIpResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFloatingIpResponse struct{}"
	}

	return strings.Join([]string{"NeutronShowFloatingIpResponse", string(data)}, " ")
}
