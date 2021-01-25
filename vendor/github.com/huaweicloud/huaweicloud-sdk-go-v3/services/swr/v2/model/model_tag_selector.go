/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TagSelector struct {
	// 匹配规则，label、regexp
	Kind string `json:"kind"`
	// kind是label时，设置为镜像版本,kind是regexp时，设置为正则表达式
	Pattern string `json:"pattern"`
}

func (o TagSelector) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagSelector struct{}"
	}

	return strings.Join([]string{"TagSelector", string(data)}, " ")
}
