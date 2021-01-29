/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 任务信息。
type GetTaskDetailListRspJobs struct {
	// 任务ID。
	Id string `json:"id"`
	// 任务名称。
	Name string `json:"name"`
	// 任务执行状态。  取值： - 值为“Running”，表示任务正在执行。 - 值为“Completed”，表示任务执行成功。 - 值为“Failed”，表示任务执行失败。
	Status GetTaskDetailListRspJobsStatus `json:"status"`
	// 创建时间，格式为“yyyy-mm-ddThh:mm:ssZ”。  其中，T指某个时间的开始；Z指时区偏移量，例如北京时间偏移显示为+0800。
	Created string `json:"created"`
	// 结束时间，格式为“yyyy-mm-ddThh:mm:ssZ”。  其中，T指某个时间的开始；Z指时区偏移量，例如北京时间偏移显示为+0800。
	Ended *string `json:"ended,omitempty"`
	// 任务执行进度。执行中状态才返回执行进度，例如60%，否则返回“”。
	Process *string `json:"process,omitempty"`
}

func (o GetTaskDetailListRspJobs) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetTaskDetailListRspJobs struct{}"
	}

	return strings.Join([]string{"GetTaskDetailListRspJobs", string(data)}, " ")
}

type GetTaskDetailListRspJobsStatus struct {
	value string
}

type GetTaskDetailListRspJobsStatusEnum struct {
	RUNNING   GetTaskDetailListRspJobsStatus
	COMPLETED GetTaskDetailListRspJobsStatus
	FAILED    GetTaskDetailListRspJobsStatus
}

func GetGetTaskDetailListRspJobsStatusEnum() GetTaskDetailListRspJobsStatusEnum {
	return GetTaskDetailListRspJobsStatusEnum{
		RUNNING: GetTaskDetailListRspJobsStatus{
			value: "Running",
		},
		COMPLETED: GetTaskDetailListRspJobsStatus{
			value: "Completed",
		},
		FAILED: GetTaskDetailListRspJobsStatus{
			value: "Failed",
		},
	}
}

func (c GetTaskDetailListRspJobsStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *GetTaskDetailListRspJobsStatus) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
