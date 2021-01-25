/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ListProductsRespValues struct {
	// 规格详情。
	Detail *[]ListProductsRespDetail `json:"detail,omitempty"`
	// 实例类型。
	Name *string `json:"name,omitempty"`
	// 资源售罄的可用区列表。
	UnavailableZones *[]string `json:"unavailable_zones,omitempty"`
	// 有可用资源的可用区列表。
	AvailableZones *[]string `json:"available_zones,omitempty"`
}

func (o ListProductsRespValues) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProductsRespValues struct{}"
	}

	return strings.Join([]string{"ListProductsRespValues", string(data)}, " ")
}
