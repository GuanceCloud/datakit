/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowSubNetworkInterfacesQuantityResponse struct {
	// 请求ID
	RequestId *string `json:"request_id,omitempty"`
	// 辅助弹性网卡数目
	SubNetworkInterfaces *int32 `json:"sub_network_interfaces,omitempty"`
	HttpStatusCode       int    `json:"-"`
}

func (o ShowSubNetworkInterfacesQuantityResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSubNetworkInterfacesQuantityResponse struct{}"
	}

	return strings.Join([]string{"ShowSubNetworkInterfacesQuantityResponse", string(data)}, " ")
}
