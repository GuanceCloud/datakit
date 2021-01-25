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
type ListBandwidthsResponse struct {
	// 带宽列表对象
	Bandwidths     *[]BandwidthResp `json:"bandwidths,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListBandwidthsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBandwidthsResponse struct{}"
	}

	return strings.Join([]string{"ListBandwidthsResponse", string(data)}, " ")
}
