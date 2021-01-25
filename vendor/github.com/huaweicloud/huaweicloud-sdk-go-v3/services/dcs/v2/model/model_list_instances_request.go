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

// Request Object
type ListInstancesRequest struct {
	InstanceId     *string `json:"instance_id,omitempty"`
	IncludeFailure *string `json:"include_failure,omitempty"`
	Name           *string `json:"name,omitempty"`
	Offset         *int32  `json:"offset,omitempty"`
	Limit          *int32  `json:"limit,omitempty"`
	Status         *string `json:"status,omitempty"`
	NameEqual      *string `json:"name_equal,omitempty"`
	Tags           *string `json:"tags,omitempty"`
	Ip             *string `json:"ip,omitempty"`
}

func (o ListInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListInstancesRequest", string(data)}, " ")
}
