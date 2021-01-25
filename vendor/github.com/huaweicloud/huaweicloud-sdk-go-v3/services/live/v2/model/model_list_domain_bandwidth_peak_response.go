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
type ListDomainBandwidthPeakResponse struct {
	// 域名对应的带宽峰值列表。
	BandwidthList  *[]PeakBandwidthData `json:"bandwidth_list,omitempty"`
	XRequestId     *string              `json:"X-request-id,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ListDomainBandwidthPeakResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainBandwidthPeakResponse struct{}"
	}

	return strings.Join([]string{"ListDomainBandwidthPeakResponse", string(data)}, " ")
}
