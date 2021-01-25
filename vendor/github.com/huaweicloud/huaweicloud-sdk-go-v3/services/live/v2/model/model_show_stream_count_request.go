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
type ShowStreamCountRequest struct {
	PublishDomains []string `json:"publish_domains"`
	StartTime      *string  `json:"start_time,omitempty"`
	EndTime        *string  `json:"end_time,omitempty"`
}

func (o ShowStreamCountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowStreamCountRequest struct{}"
	}

	return strings.Join([]string{"ShowStreamCountRequest", string(data)}, " ")
}
