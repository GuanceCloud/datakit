/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 快照参数。
type InstanceSnapshotView struct {
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 描述。
	Description *string `json:"description,omitempty"`
	// 应用组件实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 版本号。
	Version *string `json:"version,omitempty"`
}

func (o InstanceSnapshotView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceSnapshotView struct{}"
	}

	return strings.Join([]string{"InstanceSnapshotView", string(data)}, " ")
}
