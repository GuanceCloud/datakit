/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListPipleineBuildResultResponse struct {
	// 偏移量,表示从此偏移量开始查询,offset大于等于0
	Offset *int32 `json:"offset,omitempty"`
	// 每次查询的条目数量
	Limit *int32 `json:"limit,omitempty"`
	// 总条目数量
	Total *int32 `json:"total,omitempty"`
	// 执行状况数据列表
	BuildResults   *[]PipelineBuildResult `json:"build_results,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o ListPipleineBuildResultResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPipleineBuildResultResponse struct{}"
	}

	return strings.Join([]string{"ListPipleineBuildResultResponse", string(data)}, " ")
}
