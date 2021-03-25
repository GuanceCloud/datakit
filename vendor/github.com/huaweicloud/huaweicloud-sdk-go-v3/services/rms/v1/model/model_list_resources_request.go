/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListResourcesRequest struct {
	Provider string  `json:"provider"`
	Type     string  `json:"type"`
	RegionId *string `json:"region_id,omitempty"`
	EpId     *string `json:"ep_id,omitempty"`
	Limit    *int32  `json:"limit,omitempty"`
	Marker   *string `json:"marker,omitempty"`
}

func (o ListResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourcesRequest struct{}"
	}

	return strings.Join([]string{"ListResourcesRequest", string(data)}, " ")
}
