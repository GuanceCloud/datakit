/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListPrivateipsRequest struct {
	SubnetId string  `json:"subnet_id"`
	Limit    *int32  `json:"limit,omitempty"`
	Marker   *string `json:"marker,omitempty"`
}

func (o ListPrivateipsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPrivateipsRequest struct{}"
	}

	return strings.Join([]string{"ListPrivateipsRequest", string(data)}, " ")
}
