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

// 协调器信息。
type ShowCoordinatorsRespCoordinators struct {
	// 消费组ID。
	GroupId *string `json:"group_id,omitempty"`
	// 对应协调器的broker id。
	Id *int32 `json:"id,omitempty"`
	// 对应协调器的地址。
	Host *string `json:"host,omitempty"`
	// 端口号。
	Port *int32 `json:"port,omitempty"`
}

func (o ShowCoordinatorsRespCoordinators) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCoordinatorsRespCoordinators struct{}"
	}

	return strings.Join([]string{"ShowCoordinatorsRespCoordinators", string(data)}, " ")
}
