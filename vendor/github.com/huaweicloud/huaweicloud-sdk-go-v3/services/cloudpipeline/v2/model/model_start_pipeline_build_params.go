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

type StartPipelineBuildParams struct {
	// 构建参数名
	Name string `json:"name"`
	// 构建参数值，最大长度为8192
	Value string `json:"value"`
}

func (o StartPipelineBuildParams) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartPipelineBuildParams struct{}"
	}

	return strings.Join([]string{"StartPipelineBuildParams", string(data)}, " ")
}
