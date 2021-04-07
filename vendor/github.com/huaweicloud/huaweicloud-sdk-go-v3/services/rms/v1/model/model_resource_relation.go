/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ResourceRelation struct {
	// 关系类型
	RelationType *string `json:"relation_type,omitempty"`
	// 源资源类型
	FromResourceType *string `json:"from_resource_type,omitempty"`
	// 目的资源类型
	ToResourceType *string `json:"to_resource_type,omitempty"`
	// 源资源ID
	FromResourceId *string `json:"from_resource_id,omitempty"`
	// 目的资源ID
	ToResourceId *string `json:"to_resource_id,omitempty"`
}

func (o ResourceRelation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceRelation struct{}"
	}

	return strings.Join([]string{"ResourceRelation", string(data)}, " ")
}
