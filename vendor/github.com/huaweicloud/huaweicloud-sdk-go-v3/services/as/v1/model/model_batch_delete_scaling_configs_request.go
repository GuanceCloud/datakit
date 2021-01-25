/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchDeleteScalingConfigsRequest struct {
	Body *BatchDeleteScalingConfigsRequestBody `json:"body,omitempty"`
}

func (o BatchDeleteScalingConfigsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingConfigsRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingConfigsRequest", string(data)}, " ")
}
