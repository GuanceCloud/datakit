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
type CreateScalingConfigRequest struct {
	Body *CreateScalingConfigRequestBody `json:"body,omitempty"`
}

func (o CreateScalingConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateScalingConfigRequest struct{}"
	}

	return strings.Join([]string{"CreateScalingConfigRequest", string(data)}, " ")
}
