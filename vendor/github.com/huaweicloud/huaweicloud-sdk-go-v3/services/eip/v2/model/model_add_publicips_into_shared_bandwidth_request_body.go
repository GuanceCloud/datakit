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

// 将弹性公网IP插入共享带宽的请求体
type AddPublicipsIntoSharedBandwidthRequestBody struct {
	Bandwidth *AddPublicipsIntoSharedBandwidthOption `json:"bandwidth"`
}

func (o AddPublicipsIntoSharedBandwidthRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddPublicipsIntoSharedBandwidthRequestBody struct{}"
	}

	return strings.Join([]string{"AddPublicipsIntoSharedBandwidthRequestBody", string(data)}, " ")
}
