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

// Response Object
type ListInstanceSnapshotsResponse struct {
	// 快照总数。
	Count *int32 `json:"count,omitempty"`
	// 快照列表。
	Snapshots      *[]InstanceSnapshotView `json:"snapshots,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ListInstanceSnapshotsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstanceSnapshotsResponse struct{}"
	}

	return strings.Join([]string{"ListInstanceSnapshotsResponse", string(data)}, " ")
}
