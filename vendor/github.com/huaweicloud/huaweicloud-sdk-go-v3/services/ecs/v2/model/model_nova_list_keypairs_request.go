/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type NovaListKeypairsRequest struct {
	Limit               *int32  `json:"limit,omitempty"`
	Marker              *string `json:"marker,omitempty"`
	OpenStackAPIVersion *string `json:"OpenStack-API-Version,omitempty"`
}

func (o NovaListKeypairsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaListKeypairsRequest struct{}"
	}

	return strings.Join([]string{"NovaListKeypairsRequest", string(data)}, " ")
}
