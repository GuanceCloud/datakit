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
type NeutronDeleteFloatingIpResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NeutronDeleteFloatingIpResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFloatingIpResponse struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFloatingIpResponse", string(data)}, " ")
}
