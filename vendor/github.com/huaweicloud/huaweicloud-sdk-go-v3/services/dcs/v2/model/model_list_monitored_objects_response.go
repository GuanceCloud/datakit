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

// Response Object
type ListMonitoredObjectsResponse struct {
	// 当前查询维度路由。如果是主维度，则数组中是自身ID。
	Router *[]string `json:"router,omitempty"`
	// 当前查询维度子维度对象列表。当前只有维度为dcs_instance_id时才有值。 - Proxy集群有两个子维度，分别为dcs_cluster_redis_node和dcs_cluster_proxy_node。 - Cluster集群有一个子维度 dcs_cluster_proxy_node。
	Children *[]DimChild `json:"children,omitempty"`
	// 当前查询维度监控对象列表。
	Instances *[]InstancesMonitoredObject `json:"instances,omitempty"`
	// 主维度监控对象的总数。
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListMonitoredObjectsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMonitoredObjectsResponse struct{}"
	}

	return strings.Join([]string{"ListMonitoredObjectsResponse", string(data)}, " ")
}
