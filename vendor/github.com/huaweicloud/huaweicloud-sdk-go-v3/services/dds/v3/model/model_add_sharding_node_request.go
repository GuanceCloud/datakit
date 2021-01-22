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
type AddShardingNodeRequest struct {
	InstanceId string                      `json:"instance_id"`
	Body       *EnlargeInstanceRequestBody `json:"body,omitempty"`
}

func (o AddShardingNodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddShardingNodeRequest struct{}"
	}

	return strings.Join([]string{"AddShardingNodeRequest", string(data)}, " ")
}
