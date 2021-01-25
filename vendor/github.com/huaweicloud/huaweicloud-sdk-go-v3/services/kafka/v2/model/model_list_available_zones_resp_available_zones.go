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

type ListAvailableZonesRespAvailableZones struct {
	// 是否售罄。
	SoldOut *bool `json:"soldOut,omitempty"`
	// 可用区ID。
	Id *string `json:"id,omitempty"`
	// 可用区编码。
	Code *string `json:"code,omitempty"`
	// 可用区名称。
	Name *string `json:"name,omitempty"`
	// 可用区端口号。
	Port *string `json:"port,omitempty"`
	// 分区上是否还有可用资源。
	ResourceAvailability *string `json:"resource_availability,omitempty"`
	// 是否为默认可用区。
	DefaultAz *bool `json:"default_az,omitempty"`
	// 是否支持IPv6。
	Ipv6Enable *bool `json:"ipv6_enable,omitempty"`
}

func (o ListAvailableZonesRespAvailableZones) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAvailableZonesRespAvailableZones struct{}"
	}

	return strings.Join([]string{"ListAvailableZonesRespAvailableZones", string(data)}, " ")
}
