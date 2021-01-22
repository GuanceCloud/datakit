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

// 批量创建带宽的请求体
type BatchCreateBandwidthRequestBody struct {
	Bandwidth *BatchCreateBandwidthOption `json:"bandwidth"`
}

func (o BatchCreateBandwidthRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateBandwidthRequestBody struct{}"
	}

	return strings.Join([]string{"BatchCreateBandwidthRequestBody", string(data)}, " ")
}
