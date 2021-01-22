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

type Project struct {
	// 项目ID。
	Id string `json:"id"`
	// 项目名称。
	Name string `json:"name"`
	// 项目的clone url路径。
	CloneUrl string `json:"clone_url"`
}

func (o Project) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Project struct{}"
	}

	return strings.Join([]string{"Project", string(data)}, " ")
}
