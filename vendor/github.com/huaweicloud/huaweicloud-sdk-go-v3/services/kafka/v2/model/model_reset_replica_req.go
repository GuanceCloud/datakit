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

// 期望调整的分区副本分配情况。
type ResetReplicaReq struct {
	// 期望调整的分区副本分配情况。
	Partitions *[]ResetReplicaReqPartitions `json:"partitions,omitempty"`
}

func (o ResetReplicaReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetReplicaReq struct{}"
	}

	return strings.Join([]string{"ResetReplicaReq", string(data)}, " ")
}
