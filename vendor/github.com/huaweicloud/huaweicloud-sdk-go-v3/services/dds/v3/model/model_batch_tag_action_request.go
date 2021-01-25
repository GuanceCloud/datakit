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
type BatchTagActionRequest struct {
	InstanceId string                              `json:"instance_id"`
	Body       *BatchOperateInstanceTagRequestBody `json:"body,omitempty"`
}

func (o BatchTagActionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchTagActionRequest struct{}"
	}

	return strings.Join([]string{"BatchTagActionRequest", string(data)}, " ")
}
