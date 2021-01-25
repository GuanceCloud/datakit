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
type NeutronUpdateFloatingIpResponse struct {
	Floatingip     *PostAndPutFloatingIpResp `json:"floatingip,omitempty"`
	HttpStatusCode int                       `json:"-"`
}

func (o NeutronUpdateFloatingIpResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFloatingIpResponse struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFloatingIpResponse", string(data)}, " ")
}
