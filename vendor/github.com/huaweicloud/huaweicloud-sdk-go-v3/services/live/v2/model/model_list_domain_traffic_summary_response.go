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
type ListDomainTrafficSummaryResponse struct {
	// 域名对应的流量汇总列表。
	TrafficList    *[]TrafficSummaryData `json:"traffic_list,omitempty"`
	XRequestId     *string               `json:"X-request-id,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o ListDomainTrafficSummaryResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainTrafficSummaryResponse struct{}"
	}

	return strings.Join([]string{"ListDomainTrafficSummaryResponse", string(data)}, " ")
}
