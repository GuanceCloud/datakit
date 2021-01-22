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

// 更新带宽的请求体
type UpdatePrePaidBandwidthRequestBody struct {
	Bandwidth   *UpdatePrePaidBandwidthOption            `json:"bandwidth"`
	ExtendParam *UpdatePrePaidBandwidthExtendParamOption `json:"extendParam,omitempty"`
}

func (o UpdatePrePaidBandwidthRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePrePaidBandwidthRequestBody struct{}"
	}

	return strings.Join([]string{"UpdatePrePaidBandwidthRequestBody", string(data)}, " ")
}
