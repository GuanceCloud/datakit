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

type CceJobSpec struct {
	// 作业所在的集群的ID。
	ClusterUID *string `json:"clusterUID,omitempty"`
	// 扩展参数。
	ExtendParam map[string]string `json:"extendParam,omitempty"`
	// 作业操作的资源ID。
	ResourceID *string `json:"resourceID,omitempty"`
	// 作业操作的资源名称。
	ResourceName *string `json:"resourceName,omitempty"`
	// 子作业的列表。  - 包含了所有子作业的详细信息 - 在创建集群、节点等场景下，通常会由多个子作业共同组成创建作业，在子作业都完成后，作业才会完成
	SubJobs *[]CceJob `json:"subJobs,omitempty"`
	// 作业的类型，例：“CreateCluster”- 创建集群。
	Type *string `json:"type,omitempty"`
}

func (o CceJobSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CceJobSpec struct{}"
	}

	return strings.Join([]string{"CceJobSpec", string(data)}, " ")
}
