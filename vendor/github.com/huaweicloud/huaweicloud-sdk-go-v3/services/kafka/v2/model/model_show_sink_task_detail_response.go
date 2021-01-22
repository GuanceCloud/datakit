/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowSinkTaskDetailResponse struct {
	// 转储任务名称。
	TaskName *string `json:"task_name,omitempty"`
	// 转储任务类型。
	DestinationType *string `json:"destination_type,omitempty"`
	// 转储任务创建时间戳。
	CreateTime *string `json:"create_time,omitempty"`
	// 转储任务状态。
	Status *string `json:"status,omitempty"`
	// 返回任务转存的topics列表或者正则表达式。
	Topics                   *string                                         `json:"topics,omitempty"`
	ObsDestinationDescriptor *ShowSinkTaskDetailRespObsDestinationDescriptor `json:"obs_destination_descriptor,omitempty"`
	HttpStatusCode           int                                             `json:"-"`
}

func (o ShowSinkTaskDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSinkTaskDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowSinkTaskDetailResponse", string(data)}, " ")
}
