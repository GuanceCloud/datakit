/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ShowCeshierarchyRespNodes struct {
	// 节点名称。
	Name *string `json:"name,omitempty"`
}

func (o ShowCeshierarchyRespNodes) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCeshierarchyRespNodes struct{}"
	}

	return strings.Join([]string{"ShowCeshierarchyRespNodes", string(data)}, " ")
}
