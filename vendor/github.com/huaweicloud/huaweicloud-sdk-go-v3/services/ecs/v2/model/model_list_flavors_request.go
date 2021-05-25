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
type ListFlavorsRequest struct {
	AvailabilityZone *string `json:"availability_zone,omitempty"`
}

func (o ListFlavorsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFlavorsRequest struct{}"
	}

	return strings.Join([]string{"ListFlavorsRequest", string(data)}, " ")
}
