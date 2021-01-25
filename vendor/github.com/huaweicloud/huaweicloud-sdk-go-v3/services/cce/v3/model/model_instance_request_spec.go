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

// spec是集合类的元素类型，内容为插件实例安装/升级的具体请求信息
type InstanceRequestSpec struct {
	// 待安装插件模板名称，如coredns
	AddonTemplateName string `json:"addonTemplateName"`
	// 集群id
	ClusterID string `json:"clusterID"`
	// 插件模板安装参数（各插件不同）
	Values map[string]interface{} `json:"values"`
	// 待安装、升级插件的具体版本版本号，例如1.0.0
	Version string `json:"version"`
}

func (o InstanceRequestSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceRequestSpec struct{}"
	}

	return strings.Join([]string{"InstanceRequestSpec", string(data)}, " ")
}
