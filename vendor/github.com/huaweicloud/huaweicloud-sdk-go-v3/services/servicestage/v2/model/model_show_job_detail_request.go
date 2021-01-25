/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowJobDetailRequest struct {
	JobId      string  `json:"job_id"`
	InstanceId *string `json:"instance_id,omitempty"`
	Limit      *int32  `json:"limit,omitempty"`
	Offset     *int32  `json:"offset,omitempty"`
	Desc       *string `json:"desc,omitempty"`
}

func (o ShowJobDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowJobDetailRequest", string(data)}, " ")
}
