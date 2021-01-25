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

// spec是集合类的元素类型，内容为插件实例具体信息
type InstanceSpec struct {
	// 插件模板所属类型
	AddonTemplateLabels *[]string `json:"addonTemplateLabels,omitempty"`
	// 插件logo
	AddonTemplateLogo *string `json:"addonTemplateLogo,omitempty"`
	// 插件模板名称，如coredns
	AddonTemplateName string `json:"addonTemplateName"`
	// 插件模板类型
	AddonTemplateType string `json:"addonTemplateType"`
	// 集群id
	ClusterID string `json:"clusterID"`
	// 插件模板描述
	Description string `json:"description"`
	// 插件模板安装参数（各插件不同）
	Values map[string]interface{} `json:"values"`
	// 插件模板版本号，如1.0.0
	Version string `json:"version"`
}

func (o InstanceSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceSpec struct{}"
	}

	return strings.Join([]string{"InstanceSpec", string(data)}, " ")
}
