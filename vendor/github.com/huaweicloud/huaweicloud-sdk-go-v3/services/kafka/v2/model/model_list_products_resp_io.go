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

type ListProductsRespIo struct {
	// IO类型。
	IoType *string `json:"io_type,omitempty"`
	// IO规格。
	StorageSpecCode *string `json:"storage_spec_code,omitempty"`
	// IO未售罄的可用区列表。
	AvailableZones *[]string `json:"available_zones,omitempty"`
	// IO已售罄的不可用区列表。
	UnavailableZones *[]string `json:"unavailable_zones,omitempty"`
	// 磁盘类型。
	VolumeType *string `json:"volume_type,omitempty"`
}

func (o ListProductsRespIo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProductsRespIo struct{}"
	}

	return strings.Join([]string{"ListProductsRespIo", string(data)}, " ")
}
