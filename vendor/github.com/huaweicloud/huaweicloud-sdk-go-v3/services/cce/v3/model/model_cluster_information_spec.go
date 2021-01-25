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

type ClusterInformationSpec struct {
	// 集群的描述信息。  1. 字符取值范围[0,200]。不包含~$%^&*<>[]{}()'\"#\\等特殊字符。 2. 仅运行和扩容状态（Available、ScalingUp、ScalingDown）的集群允许修改。
	Description *string `json:"description,omitempty"`
}

func (o ClusterInformationSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClusterInformationSpec struct{}"
	}

	return strings.Join([]string{"ClusterInformationSpec", string(data)}, " ")
}
