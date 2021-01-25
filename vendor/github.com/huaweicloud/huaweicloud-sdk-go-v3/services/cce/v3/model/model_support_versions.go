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

// 插件支持升级的集群版本
type SupportVersions struct {
	// 支持的集群类型
	ClusterType string `json:"clusterType"`
	// 支持的集群版本（正则表达式）
	ClusterVersion []string `json:"clusterVersion"`
}

func (o SupportVersions) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SupportVersions struct{}"
	}

	return strings.Join([]string{"SupportVersions", string(data)}, " ")
}
