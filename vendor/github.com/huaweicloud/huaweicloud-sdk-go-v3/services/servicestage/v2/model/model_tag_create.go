/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TagCreate struct {
	// tag标签名称。
	Name string `json:"name"`
	// tag标签描述信息。
	Description string `json:"description"`
}

func (o TagCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagCreate struct{}"
	}

	return strings.Join([]string{"TagCreate", string(data)}, " ")
}
