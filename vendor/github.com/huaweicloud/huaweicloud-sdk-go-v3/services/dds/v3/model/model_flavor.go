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

// 规格信息。
type Flavor struct {
	// 引擎名称。
	EngineName string `json:"engine_name"`
	// 节点类型。文档数据库包含以下几种节点类型： - mongos - shard - config - replica - single
	Type string `json:"type"`
	// CPU核数。
	Vcpus string `json:"vcpus"`
	// 内存大小，单位为兆字节。
	Ram string `json:"ram"`
	// 资源规格编码。例如：dds.c3.xlarge.2.shard。  - “dds”表示文档数据库服务产品。 - “c3.xlarge.2”表示节点性能规格，为高内存类型。 - “shard”表示节点类型。
	SpecCode string `json:"spec_code"`
	// '支持该规格的可用区ID。' 示例：[\"cn-east-2a\",\"cn-east-2b\",\"cn-east-2c\"]。
	AzStatus *interface{} `json:"az_status"`
}

func (o Flavor) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Flavor struct{}"
	}

	return strings.Join([]string{"Flavor", string(data)}, " ")
}
