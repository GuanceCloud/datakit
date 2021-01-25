/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListInstanceTagsRequest struct {
	InstanceId string `json:"instance_id"`
}

func (o ListInstanceTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstanceTagsRequest struct{}"
	}

	return strings.Join([]string{"ListInstanceTagsRequest", string(data)}, " ")
}
