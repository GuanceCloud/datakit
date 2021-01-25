/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 实例节点信息。
type NodeItem struct {
	// 节点ID。
	Id string `json:"id"`
	// 节点名称。
	Name string `json:"name"`
	// 节点状态。
	Status string `json:"status"`
	// 节点角色。 取值： - master，mongos节点返回该值。 - Primary，shard组主节点、config组主节点、副本集主节点、单节点返回该值。 - Secondary，shard组备节点、config组备节点、副本集备节点返回该值。 - Hidden，shard组隐藏节点、config组隐藏节点、副本集隐藏节点返回该值。 - unknown，节点异常时返回该值。
	Role string `json:"role"`
	// 节点内网IP。该参数仅针对集群实例的mongos节点、副本集实例、以及单节点实例有效，且在弹性云服务器创建成功后参数值存在，否则，值为\"\"。
	PrivateIp string `json:"private_ip"`
	// 绑定的外网IP。该参数值为\"\"。该参数仅针对集群实例的mongos节点、副本集实例的主节点和备节点、以及单节点实例有效。
	PublicIp string `json:"public_ip"`
	// 资源规格编码。
	SpecCode string `json:"spec_code"`
	// 可用区。
	AvailabilityZone string `json:"availability_zone"`
}

func (o NodeItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NodeItem struct{}"
	}

	return strings.Join([]string{"NodeItem", string(data)}, " ")
}
