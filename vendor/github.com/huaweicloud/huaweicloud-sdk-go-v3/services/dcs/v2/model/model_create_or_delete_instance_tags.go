/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateOrDeleteInstanceTags struct {
	// 操作标识：仅限于create（创建）、delete（删除）。
	Action string `json:"action"`
	// 标签列表。
	Tags *[]ResourceTag `json:"tags,omitempty"`
}

func (o CreateOrDeleteInstanceTags) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateOrDeleteInstanceTags struct{}"
	}

	return strings.Join([]string{"CreateOrDeleteInstanceTags", string(data)}, " ")
}
