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

// Request Object
type ListDomainTrafficSummaryRequest struct {
	PlayDomains []string  `json:"play_domains"`
	App         *string   `json:"app,omitempty"`
	Stream      *string   `json:"stream,omitempty"`
	Region      *[]string `json:"region,omitempty"`
	Isp         *[]string `json:"isp,omitempty"`
	StartTime   *string   `json:"start_time,omitempty"`
	EndTime     *string   `json:"end_time,omitempty"`
}

func (o ListDomainTrafficSummaryRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainTrafficSummaryRequest struct{}"
	}

	return strings.Join([]string{"ListDomainTrafficSummaryRequest", string(data)}, " ")
}
