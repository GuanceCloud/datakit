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
type RemovePublicipsFromSharedBandwidthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o RemovePublicipsFromSharedBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemovePublicipsFromSharedBandwidthResponse struct{}"
	}

	return strings.Join([]string{"RemovePublicipsFromSharedBandwidthResponse", string(data)}, " ")
}
