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
type ListAllResourcesRequest struct {
	RegionId *string `json:"region_id,omitempty"`
	EpId     *string `json:"ep_id,omitempty"`
	Type     *string `json:"type,omitempty"`
	Limit    *int32  `json:"limit,omitempty"`
	Marker   *string `json:"marker,omitempty"`
}

func (o ListAllResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAllResourcesRequest struct{}"
	}

	return strings.Join([]string{"ListAllResourcesRequest", string(data)}, " ")
}
