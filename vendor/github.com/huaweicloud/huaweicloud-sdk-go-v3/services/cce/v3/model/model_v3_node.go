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

type V3Node struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“Node”，该值不可修改。
	Kind     *string       `json:"kind,omitempty"`
	Metadata *NodeMetadata `json:"metadata,omitempty"`
	Spec     *V3NodeSpec   `json:"spec,omitempty"`
	Status   *V3NodeStatus `json:"status,omitempty"`
}

func (o V3Node) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V3Node struct{}"
	}

	return strings.Join([]string{"V3Node", string(data)}, " ")
}
