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
type ListAvailableZonesResponse struct {
	// 区域ID。
	RegionId *string `json:"region_id,omitempty"`
	// 可用分区数组。
	AvailableZones *[]AvailableZones `json:"available_zones,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListAvailableZonesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAvailableZonesResponse struct{}"
	}

	return strings.Join([]string{"ListAvailableZonesResponse", string(data)}, " ")
}
