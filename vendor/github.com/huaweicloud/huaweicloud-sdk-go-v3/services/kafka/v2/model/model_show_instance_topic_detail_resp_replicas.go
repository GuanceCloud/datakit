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

type ShowInstanceTopicDetailRespReplicas struct {
	// 副本所在的节点ID。
	Broker *int32 `json:"broker,omitempty"`
	// 该副本是否为leader。
	Leader *bool `json:"leader,omitempty"`
	// 该副本是否在ISR副本中。
	InSync *bool `json:"in_sync,omitempty"`
	// 该副本当前日志大小。
	Size *int32 `json:"size,omitempty"`
	// 该副本当前落后hw的消息数。
	Lag *int32 `json:"lag,omitempty"`
}

func (o ShowInstanceTopicDetailRespReplicas) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceTopicDetailRespReplicas struct{}"
	}

	return strings.Join([]string{"ShowInstanceTopicDetailRespReplicas", string(data)}, " ")
}
