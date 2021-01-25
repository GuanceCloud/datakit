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

type ListProductsRespDetail struct {
	// 单位时间内的消息量最大值。
	Tps *string `json:"tps,omitempty"`
	// 消息存储空间。
	Storage *string `json:"storage,omitempty"`
	// Kafka实例的最大Topic数。
	PartitionNum *string `json:"partition_num,omitempty"`
	// 产品ID。
	ProductId *string `json:"product_id,omitempty"`
	// 规格ID。
	SpecCode *string `json:"spec_code,omitempty"`
	// IO信息。
	Io *[]ListProductsRespIo `json:"io,omitempty"`
	// Kafka实例的基准带宽。
	Bandwidth *string `json:"bandwidth,omitempty"`
	// 资源售罄的可用区列表。
	UnavailableZones *[]string `json:"unavailable_zones,omitempty"`
	// 有可用资源的可用区列表。
	AvailableZones *[]string `json:"available_zones,omitempty"`
	// 该产品规格对应的虚拟机规格。
	EcsFlavorId *string `json:"ecs_flavor_id,omitempty"`
	// 实例规格架构类型。当前仅支持X86。
	ArchType *string `json:"arch_type,omitempty"`
}

func (o ListProductsRespDetail) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProductsRespDetail struct{}"
	}

	return strings.Join([]string{"ListProductsRespDetail", string(data)}, " ")
}
