/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 监控对象结构
type ClusterRedisNodeMonitoredObject struct {
	// 测量对象ID，即节点的ID。
	DcsInstanceId *string `json:"dcs_instance_id,omitempty"`
	// 测量对象名称，即节点IP。
	Name *string `json:"name,omitempty"`
	// 维度dcs_cluster_redis_node的测量对象的ID。
	DcsClusterRedisNode *string `json:"dcs_cluster_redis_node,omitempty"`
	// 测量对象状态，即节点状态。
	Status *string `json:"status,omitempty"`
}

func (o ClusterRedisNodeMonitoredObject) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClusterRedisNodeMonitoredObject struct{}"
	}

	return strings.Join([]string{"ClusterRedisNodeMonitoredObject", string(data)}, " ")
}
