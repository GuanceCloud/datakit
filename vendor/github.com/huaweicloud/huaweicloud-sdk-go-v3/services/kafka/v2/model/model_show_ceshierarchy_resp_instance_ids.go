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

type ShowCeshierarchyRespInstanceIds struct {
	// 实例ID。
	Name *string `json:"name,omitempty"`
}

func (o ShowCeshierarchyRespInstanceIds) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCeshierarchyRespInstanceIds struct{}"
	}

	return strings.Join([]string{"ShowCeshierarchyRespInstanceIds", string(data)}, " ")
}
