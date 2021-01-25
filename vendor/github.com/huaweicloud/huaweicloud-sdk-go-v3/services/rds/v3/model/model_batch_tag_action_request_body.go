/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type BatchTagActionRequestBody struct {
	// 操作标识（区分大小写）：create（创建）、delete（删除）。
	Action string `json:"action"`
	// 标签列表。单个实例总标签数上限10个。
	Tags []InstanceRequestTags `json:"tags"`
}

func (o BatchTagActionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchTagActionRequestBody struct{}"
	}

	return strings.Join([]string{"BatchTagActionRequestBody", string(data)}, " ")
}
