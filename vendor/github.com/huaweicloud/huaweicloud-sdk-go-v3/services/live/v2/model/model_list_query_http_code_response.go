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
type ListQueryHttpCodeResponse struct {
	// 基于时间轴的状态码
	DataSeries     *[]HttpCodeSummary `json:"data_series,omitempty"`
	XRequestId     *string            `json:"X-request-id,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListQueryHttpCodeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQueryHttpCodeResponse struct{}"
	}

	return strings.Join([]string{"ListQueryHttpCodeResponse", string(data)}, " ")
}
