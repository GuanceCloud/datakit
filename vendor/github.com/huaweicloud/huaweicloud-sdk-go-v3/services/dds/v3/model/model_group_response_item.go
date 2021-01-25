/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 实例组信息。
type GroupResponseItem struct {
	// 节点类型。 取值： - shard - config - mongos - replica - single
	Type string `json:"type"`
	// 组ID。节点类型为shard和config时，该参数有效。
	Id string `json:"id"`
	// 组名组名称。节点类型为shard和config时，该参数有效。
	Name string `json:"name"`
	// 组状态。节点类型为shard和config时，该参数有效。
	Status string  `json:"status"`
	Volume *Volume `json:"volume"`
	// 节点信息。
	Nodes []NodeItem `json:"nodes"`
}

func (o GroupResponseItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GroupResponseItem struct{}"
	}

	return strings.Join([]string{"GroupResponseItem", string(data)}, " ")
}
