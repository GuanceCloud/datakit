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
type ListDomainTrafficDetailResponse struct {
	// 采样数据列表。
	DataList       *[]TrafficData `json:"data_list,omitempty"`
	XRequestId     *string        `json:"X-request-id,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListDomainTrafficDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainTrafficDetailResponse struct{}"
	}

	return strings.Join([]string{"ListDomainTrafficDetailResponse", string(data)}, " ")
}
