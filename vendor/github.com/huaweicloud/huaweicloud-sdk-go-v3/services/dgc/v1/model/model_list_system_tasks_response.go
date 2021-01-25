/*
 * DGC
 *
 * 数据湖治理中心DGC是具有数据全生命周期管理、智能数据管理能力的一站式治理运营平台，支持行业知识库智能化建设，支持大数据存储、大数据计算分析引擎等数据底座，帮助企业快速构建从数据接入到数据分析的端到端智能数据系统，消除数据孤岛，统一数据标准，加快数据变现，实现数字化转型
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Response Object
type ListSystemTasksResponse struct {
	Id             *string                        `json:"id,omitempty"`
	Name           *string                        `json:"name,omitempty"`
	StartTime      *string                        `json:"startTime,omitempty"`
	EndTime        *string                        `json:"endTime,omitempty"`
	LastUpdate     *string                        `json:"lastUpdate,omitempty"`
	Status         *ListSystemTasksResponseStatus `json:"status,omitempty"`
	Message        *string                        `json:"message,omitempty"`
	SubTasks       *[]SubTaskStatus               `json:"subTasks,omitempty"`
	HttpStatusCode int                            `json:"-"`
}

func (o ListSystemTasksResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSystemTasksResponse struct{}"
	}

	return strings.Join([]string{"ListSystemTasksResponse", string(data)}, " ")
}

type ListSystemTasksResponseStatus struct {
	value string
}

type ListSystemTasksResponseStatusEnum struct {
	RUNNING    ListSystemTasksResponseStatus
	SUCCESSFUL ListSystemTasksResponseStatus
	FAILED     ListSystemTasksResponseStatus
}

func GetListSystemTasksResponseStatusEnum() ListSystemTasksResponseStatusEnum {
	return ListSystemTasksResponseStatusEnum{
		RUNNING: ListSystemTasksResponseStatus{
			value: "RUNNING",
		},
		SUCCESSFUL: ListSystemTasksResponseStatus{
			value: "SUCCESSFUL",
		},
		FAILED: ListSystemTasksResponseStatus{
			value: "FAILED",
		},
	}
}

func (c ListSystemTasksResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListSystemTasksResponseStatus) UnmarshalJSON(b []byte) error {
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
