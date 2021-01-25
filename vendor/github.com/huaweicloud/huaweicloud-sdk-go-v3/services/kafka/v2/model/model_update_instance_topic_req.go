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

// 修改的topic列表。
type UpdateInstanceTopicReq struct {
	// 修改的topic列表。
	Topics *[]UpdateInstanceTopicReqTopics `json:"topics,omitempty"`
}

func (o UpdateInstanceTopicReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceTopicReq struct{}"
	}

	return strings.Join([]string{"UpdateInstanceTopicReq", string(data)}, " ")
}
