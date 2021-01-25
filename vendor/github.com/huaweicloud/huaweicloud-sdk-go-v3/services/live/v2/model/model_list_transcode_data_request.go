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
type ListTranscodeDataRequest struct {
	PublishDomain *string `json:"publish_domain,omitempty"`
	StartTime     *string `json:"start_time,omitempty"`
	EndTime       *string `json:"end_time,omitempty"`
}

func (o ListTranscodeDataRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTranscodeDataRequest struct{}"
	}

	return strings.Join([]string{"ListTranscodeDataRequest", string(data)}, " ")
}
