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

type NodePool struct {
	// API版本，固定值“v3”。
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“NodePool”。
	Kind     string            `json:"kind"`
	Metadata *NodePoolMetadata `json:"metadata"`
	Spec     *NodePoolSpec     `json:"spec"`
	Status   *NodePoolStatus   `json:"status,omitempty"`
}

func (o NodePool) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NodePool struct{}"
	}

	return strings.Join([]string{"NodePool", string(data)}, " ")
}
