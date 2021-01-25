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

// Response Object
type ShowCesHierarchyResponse struct {
	// 监控维度。
	Dimensions *[]ShowCeshierarchyRespDimensions `json:"dimensions,omitempty"`
	// 实例信息。
	InstanceIds *[]ShowCeshierarchyRespInstanceIds `json:"instance_ids,omitempty"`
	// 节点信息。
	Nodes *[]ShowCeshierarchyRespNodes `json:"nodes,omitempty"`
	// 队列信息。
	Queues *[]ShowCeshierarchyRespQueues `json:"queues,omitempty"`
	// 消费组信息。
	Groups         *[]string `json:"groups,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ShowCesHierarchyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCesHierarchyResponse struct{}"
	}

	return strings.Join([]string{"ShowCesHierarchyResponse", string(data)}, " ")
}
