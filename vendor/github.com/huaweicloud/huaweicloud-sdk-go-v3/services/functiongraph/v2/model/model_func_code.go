/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// FuncCode结构返回体。
type FuncCode struct {
	// 函数代码，当CodeTye为inline/zip/jar时必选，且代码必须要进行base64编码。
	File string `json:"file"`
	// 函数代码链接。
	Link string `json:"link"`
}

func (o FuncCode) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FuncCode struct{}"
	}

	return strings.Join([]string{"FuncCode", string(data)}, " ")
}
