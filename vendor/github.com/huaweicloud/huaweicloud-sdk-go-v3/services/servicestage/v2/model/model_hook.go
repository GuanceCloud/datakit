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

type Hook struct {
	// hook ID。
	Id string `json:"id"`
	// hook类型。
	Type string `json:"type"`
	// 回滚URL。
	CallbackUrl string `json:"callback_url"`
}

func (o Hook) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Hook struct{}"
	}

	return strings.Join([]string{"Hook", string(data)}, " ")
}
