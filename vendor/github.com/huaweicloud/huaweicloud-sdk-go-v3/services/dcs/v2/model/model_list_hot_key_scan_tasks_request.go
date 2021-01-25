/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Request Object
type ListHotKeyScanTasksRequest struct {
	InstanceId string                            `json:"instance_id"`
	Offset     *int32                            `json:"offset,omitempty"`
	Limit      *int32                            `json:"limit,omitempty"`
	Status     *ListHotKeyScanTasksRequestStatus `json:"status,omitempty"`
}

func (o ListHotKeyScanTasksRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHotKeyScanTasksRequest struct{}"
	}

	return strings.Join([]string{"ListHotKeyScanTasksRequest", string(data)}, " ")
}

type ListHotKeyScanTasksRequestStatus struct {
	value string
}

type ListHotKeyScanTasksRequestStatusEnum struct {
	WAITING ListHotKeyScanTasksRequestStatus
	RUNNING ListHotKeyScanTasksRequestStatus
	SUCCESS ListHotKeyScanTasksRequestStatus
	FAILED  ListHotKeyScanTasksRequestStatus
}

func GetListHotKeyScanTasksRequestStatusEnum() ListHotKeyScanTasksRequestStatusEnum {
	return ListHotKeyScanTasksRequestStatusEnum{
		WAITING: ListHotKeyScanTasksRequestStatus{
			value: "waiting",
		},
		RUNNING: ListHotKeyScanTasksRequestStatus{
			value: "running",
		},
		SUCCESS: ListHotKeyScanTasksRequestStatus{
			value: "success",
		},
		FAILED: ListHotKeyScanTasksRequestStatus{
			value: "failed",
		},
	}
}

func (c ListHotKeyScanTasksRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListHotKeyScanTasksRequestStatus) UnmarshalJSON(b []byte) error {
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
