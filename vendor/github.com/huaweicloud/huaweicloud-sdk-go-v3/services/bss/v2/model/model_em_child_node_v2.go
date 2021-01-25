/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type EmChildNodeV2 struct {
	// |参数名称：实体关系ID| |参数约束及描述：实体关系ID|
	RelationId *string `json:"relation_id,omitempty"`
	// |参数名称：节点ID| |参数约束及描述：节点ID|
	Id *string `json:"id,omitempty"`
	// |参数名称：节点名称| |参数约束及描述：节点名称|
	Name *string `json:"name,omitempty"`
	// |参数名称：子节点列表| |参数约束以及描述：子节点列表|
	ChildNodes *[]EmChildNodeV2 `json:"child_nodes,omitempty"`
}

func (o EmChildNodeV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EmChildNodeV2 struct{}"
	}

	return strings.Join([]string{"EmChildNodeV2", string(data)}, " ")
}
