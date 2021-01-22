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

// 维度对象结构体
type DimChild struct {
	// 维度名称，当前支持维度有dcs_instance_id、dcs_cluster_redis_node、 dcs_cluster_proxy_node和dcs_memcached_instance_id。
	DimName *string `json:"dim_name,omitempty"`
	// 维度的路由，结构为主维度名称,当前维度名称，比如： dim_name字段为dcs_cluster_redis_node时，这个字段的值为dcs_instance_id,dcs_cluster_redis_node。
	DimRoute *string `json:"dim_route,omitempty"`
}

func (o DimChild) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DimChild struct{}"
	}

	return strings.Join([]string{"DimChild", string(data)}, " ")
}
