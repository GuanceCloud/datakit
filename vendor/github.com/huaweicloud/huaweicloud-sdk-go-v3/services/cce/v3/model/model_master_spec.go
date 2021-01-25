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

// master的配置，支持指定可用区、规格和故障域。若指定故障域，则必须所有master节点都需要指定故障字段。
type MasterSpec struct {
	// 可用区
	AvailabilityZone *string `json:"availabilityZone,omitempty"`
}

func (o MasterSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MasterSpec struct{}"
	}

	return strings.Join([]string{"MasterSpec", string(data)}, " ")
}
