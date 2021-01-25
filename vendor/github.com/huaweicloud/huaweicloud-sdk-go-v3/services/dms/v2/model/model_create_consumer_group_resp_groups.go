/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateConsumerGroupRespGroups struct {
	// 消费组的ID。
	Id *string `json:"id,omitempty"`
	// 消费组的名称。
	Name *string `json:"name,omitempty"`
}

func (o CreateConsumerGroupRespGroups) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateConsumerGroupRespGroups struct{}"
	}

	return strings.Join([]string{"CreateConsumerGroupRespGroups", string(data)}, " ")
}
