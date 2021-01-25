/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowJobDetailResponse struct {
	// 任务的id
	Id *string `json:"id,omitempty"`
	// 任务的名称
	Name *string `json:"name,omitempty"`
	// 任务的状态
	JobStatus *interface{} `json:"job_status,omitempty"`
	// 任务结果信息
	JobResult      *string `json:"job_result,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowJobDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowJobDetailResponse", string(data)}, " ")
}
