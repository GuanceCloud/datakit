/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 资源结构体。
type TagResource struct {
	// 资源ID。
	ResourceId     string          `json:"resource_id"`
	ResourceDetail *ResourceDetail `json:"resource_detail"`
	// 标签列表。
	Tags []ResourceTag `json:"tags"`
	// 资源名称。
	ResourceName string `json:"resource_name"`
}

func (o TagResource) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagResource struct{}"
	}

	return strings.Join([]string{"TagResource", string(data)}, " ")
}
