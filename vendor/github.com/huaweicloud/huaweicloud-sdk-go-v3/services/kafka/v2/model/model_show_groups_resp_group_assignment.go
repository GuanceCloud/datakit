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

type ShowGroupsRespGroupAssignment struct {
	// topic名称。
	Topic *string `json:"topic,omitempty"`
	// 分区列表。
	Partitions *[]string `json:"partitions,omitempty"`
}

func (o ShowGroupsRespGroupAssignment) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowGroupsRespGroupAssignment struct{}"
	}

	return strings.Join([]string{"ShowGroupsRespGroupAssignment", string(data)}, " ")
}
