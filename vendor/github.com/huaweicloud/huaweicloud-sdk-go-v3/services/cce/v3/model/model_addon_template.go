/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 插件模板详情-response结构体
type AddonTemplate struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“Addon”，该值不可修改。
	Kind     string        `json:"kind"`
	Metadata *Metadata     `json:"metadata"`
	Spec     *Templatespec `json:"spec"`
}

func (o AddonTemplate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddonTemplate struct{}"
	}

	return strings.Join([]string{"AddonTemplate", string(data)}, " ")
}
