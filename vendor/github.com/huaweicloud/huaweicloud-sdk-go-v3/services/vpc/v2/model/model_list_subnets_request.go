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
type ListSubnetsRequest struct {
	Limit  *int32  `json:"limit,omitempty"`
	Marker *string `json:"marker,omitempty"`
	VpcId  *string `json:"vpc_id,omitempty"`
	Scope  *string `json:"scope,omitempty"`
}

func (o ListSubnetsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubnetsRequest struct{}"
	}

	return strings.Join([]string{"ListSubnetsRequest", string(data)}, " ")
}
