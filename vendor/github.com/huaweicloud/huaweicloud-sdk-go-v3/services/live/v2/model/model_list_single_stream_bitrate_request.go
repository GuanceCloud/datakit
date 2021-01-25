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
type ListSingleStreamBitrateRequest struct {
	Domain    string  `json:"domain"`
	App       string  `json:"app"`
	Stream    string  `json:"stream"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
}

func (o ListSingleStreamBitrateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSingleStreamBitrateRequest struct{}"
	}

	return strings.Join([]string{"ListSingleStreamBitrateRequest", string(data)}, " ")
}
