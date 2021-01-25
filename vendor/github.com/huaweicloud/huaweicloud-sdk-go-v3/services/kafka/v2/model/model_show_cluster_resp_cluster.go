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

// 集群基本信息。
type ShowClusterRespCluster struct {
	// 控制器ID。
	Controller *string `json:"controller,omitempty"`
	// 节点列表。
	Brokers *[]ShowClusterRespClusterBrokers `json:"brokers,omitempty"`
	// 主题数量。
	TopicsCount *int32 `json:"topics_count,omitempty"`
	// 分区数量。
	PartitionsCount *int32 `json:"partitions_count,omitempty"`
	// 在线分区数量。
	OnlinePartitionsCount *int32 `json:"online_partitions_count,omitempty"`
	// 副本数量。
	ReplicasCount *int32 `json:"replicas_count,omitempty"`
	// ISR（In-Sync Replicas） 副本总数。
	IsrReplicasCount *int32 `json:"isr_replicas_count,omitempty"`
	// 消费组数量。
	ConsumersCount *int32 `json:"consumers_count,omitempty"`
}

func (o ShowClusterRespCluster) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowClusterRespCluster struct{}"
	}

	return strings.Join([]string{"ShowClusterRespCluster", string(data)}, " ")
}
