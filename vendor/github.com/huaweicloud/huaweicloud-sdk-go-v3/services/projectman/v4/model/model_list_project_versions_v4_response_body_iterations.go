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

type ListProjectVersionsV4ResponseBodyIterations struct {
	// 迭代描述
	Description *string `json:"description,omitempty"`
	// 迭代结束时间
	EndTime *string `json:"end_time,omitempty"`
	// 迭代id
	Id *int32 `json:"id,omitempty"`
	// 迭代标题
	Name *string `json:"name,omitempty"`
	// 迭代开始时间
	BeginTime *string `json:"begin_time,omitempty"`
}

func (o ListProjectVersionsV4ResponseBodyIterations) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectVersionsV4ResponseBodyIterations struct{}"
	}

	return strings.Join([]string{"ListProjectVersionsV4ResponseBodyIterations", string(data)}, " ")
}
