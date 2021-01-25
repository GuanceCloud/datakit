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

type AttrsObject struct {
	// 缓存容量。
	Capacity *string `json:"capacity,omitempty"`
	// 额外信息名，取值范围如下： - sharding_num：该规格实例支持的分片数。 - proxy_num：该规格Proxy实例支持的Proxy节点数量。如果不是Proxy实例，该参数为0。 - db_number：该规格实例的DB数量。 - max_memory：实际可使用的最大内存。 - max_connections：该规格支持的最大连接数。 - max_clients：该规格支持的最大客户端数，一般等于最大连接数。 - max_bandwidth：该规格支持的最大带宽。 - max_in_bandwidth：该规格支持的最大接入带宽，一般等于最大带宽。
	Name *string `json:"name,omitempty"`
	// 额外信息值。
	Value *string `json:"value,omitempty"`
}

func (o AttrsObject) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttrsObject struct{}"
	}

	return strings.Join([]string{"AttrsObject", string(data)}, " ")
}
