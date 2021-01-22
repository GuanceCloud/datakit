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

// 共享带宽移除弹性公网IP的请求体
type RemovePublicipsFromSharedBandwidthRequestBody struct {
	Bandwidth *RemoveFromSharedBandwidthOption `json:"bandwidth"`
}

func (o RemovePublicipsFromSharedBandwidthRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemovePublicipsFromSharedBandwidthRequestBody struct{}"
	}

	return strings.Join([]string{"RemovePublicipsFromSharedBandwidthRequestBody", string(data)}, " ")
}
