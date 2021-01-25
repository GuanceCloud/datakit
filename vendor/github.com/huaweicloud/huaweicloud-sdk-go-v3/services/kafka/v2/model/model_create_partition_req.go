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

type CreatePartitionReq struct {
	// 期望调整分区后的数量，必须大于当前分区数量，小于等于20。
	Partition *int32 `json:"partition,omitempty"`
}

func (o CreatePartitionReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePartitionReq struct{}"
	}

	return strings.Join([]string{"CreatePartitionReq", string(data)}, " ")
}
