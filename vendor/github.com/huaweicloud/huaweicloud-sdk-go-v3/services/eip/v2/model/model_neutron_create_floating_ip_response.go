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
type NeutronCreateFloatingIpResponse struct {
	Floatingip     *PostAndPutFloatingIpResp `json:"floatingip,omitempty"`
	HttpStatusCode int                       `json:"-"`
}

func (o NeutronCreateFloatingIpResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFloatingIpResponse struct{}"
	}

	return strings.Join([]string{"NeutronCreateFloatingIpResponse", string(data)}, " ")
}
