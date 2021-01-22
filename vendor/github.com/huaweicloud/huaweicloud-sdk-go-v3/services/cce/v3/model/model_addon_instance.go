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

// 插件实例详细信息-response结构体
type AddonInstance struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“Addon”，该值不可修改。
	Kind     string               `json:"kind"`
	Metadata *Metadata            `json:"metadata,omitempty"`
	Spec     *InstanceSpec        `json:"spec"`
	Status   *AddonInstanceStatus `json:"status"`
}

func (o AddonInstance) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddonInstance struct{}"
	}

	return strings.Join([]string{"AddonInstance", string(data)}, " ")
}
