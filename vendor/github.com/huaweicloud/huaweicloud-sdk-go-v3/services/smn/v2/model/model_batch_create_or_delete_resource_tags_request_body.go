/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type BatchCreateOrDeleteResourceTagsRequestBody struct {
	// 标签列表，结构体说明请参见表1。删除时tags结构体不能缺失，key不能为空或空字符串，且不针对字符集范围进行校验。
	Tags []ResourceTag `json:"tags"`
	// 操作标识：仅限于create（创建）、delete（删除）。
	Action string `json:"action"`
}

func (o BatchCreateOrDeleteResourceTagsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteResourceTagsRequestBody struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteResourceTagsRequestBody", string(data)}, " ")
}
