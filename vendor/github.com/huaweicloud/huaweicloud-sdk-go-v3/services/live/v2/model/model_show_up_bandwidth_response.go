/*
 * Live
 *
 * 数据分析服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowUpBandwidthResponse struct {
	// 采样数据列表
	DataList       *[]V2BandwidthData `json:"data_list,omitempty"`
	XRequestId     *string            `json:"X-request-id,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ShowUpBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowUpBandwidthResponse struct{}"
	}

	return strings.Join([]string{"ShowUpBandwidthResponse", string(data)}, " ")
}
