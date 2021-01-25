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

// 插件基本信息，集合类的元素类型，包含一组由不同名称定义的属性。
type Metadata struct {
	// 插件注解，由key/value组成 - 安装：固定值为{\"addon.install/type\":\"install\"} - 升级：固定值为{\"addon.upgrade/type\":\"upgrade\"}
	Annotations map[string]string `json:"annotations,omitempty"`
	// 创建时间
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
	// 插件标签，key/value对格式
	Labels map[string]string `json:"labels,omitempty"`
	// 插件名称
	Name string `json:"name"`
	// 唯一id标识
	Uid *string `json:"uid,omitempty"`
	// 更新时间
	UpdateTimestamp string `json:"updateTimestamp,omitempty"`
}

func (o Metadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Metadata struct{}"
	}

	return strings.Join([]string{"Metadata", string(data)}, " ")
}
