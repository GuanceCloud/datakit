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
type DeleteSharedBandwidthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteSharedBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSharedBandwidthResponse struct{}"
	}

	return strings.Join([]string{"DeleteSharedBandwidthResponse", string(data)}, " ")
}
