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
type ListHistoryStreamsRequest struct {
	Domain string  `json:"domain"`
	App    *string `json:"app,omitempty"`
	Offset *int32  `json:"offset,omitempty"`
	Limit  *int32  `json:"limit,omitempty"`
}

func (o ListHistoryStreamsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHistoryStreamsRequest struct{}"
	}

	return strings.Join([]string{"ListHistoryStreamsRequest", string(data)}, " ")
}
