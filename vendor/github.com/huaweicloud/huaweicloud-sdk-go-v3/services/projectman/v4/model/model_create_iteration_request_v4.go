/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateIterationRequestV4 struct {
	// 开始时间，年-月-日
	BeginTime string `json:"begin_time"`
	// 描述
	Description *string `json:"description,omitempty"`
	// 结束时间，年-月-日
	EndTime string `json:"end_time"`
	// 标题
	Name string `json:"name"`
}

func (o CreateIterationRequestV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateIterationRequestV4 struct{}"
	}

	return strings.Join([]string{"CreateIterationRequestV4", string(data)}, " ")
}
