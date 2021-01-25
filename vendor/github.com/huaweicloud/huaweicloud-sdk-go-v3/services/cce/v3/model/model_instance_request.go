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

// 插件安装/升级-request结构体
type InstanceRequest struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“Addon”，该值不可修改。
	Kind     string               `json:"kind"`
	Metadata *Metadata            `json:"metadata"`
	Spec     *InstanceRequestSpec `json:"spec"`
}

func (o InstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceRequest struct{}"
	}

	return strings.Join([]string{"InstanceRequest", string(data)}, " ")
}
