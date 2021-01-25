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
type ListBigkeyScanTasksRequest struct {
	InstanceId string                            `json:"instance_id"`
	Offset     *int32                            `json:"offset,omitempty"`
	Limit      *int32                            `json:"limit,omitempty"`
	Status     *ListBigkeyScanTasksRequestStatus `json:"status,omitempty"`
}

func (o ListBigkeyScanTasksRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBigkeyScanTasksRequest struct{}"
	}

	return strings.Join([]string{"ListBigkeyScanTasksRequest", string(data)}, " ")
}

type ListBigkeyScanTasksRequestStatus struct {
	value string
}

type ListBigkeyScanTasksRequestStatusEnum struct {
	WAITING ListBigkeyScanTasksRequestStatus
	RUNNING ListBigkeyScanTasksRequestStatus
	SUCCESS ListBigkeyScanTasksRequestStatus
	FAILED  ListBigkeyScanTasksRequestStatus
}

func GetListBigkeyScanTasksRequestStatusEnum() ListBigkeyScanTasksRequestStatusEnum {
	return ListBigkeyScanTasksRequestStatusEnum{
		WAITING: ListBigkeyScanTasksRequestStatus{
			value: "waiting",
		},
		RUNNING: ListBigkeyScanTasksRequestStatus{
			value: "running",
		},
		SUCCESS: ListBigkeyScanTasksRequestStatus{
			value: "success",
		},
		FAILED: ListBigkeyScanTasksRequestStatus{
			value: "failed",
		},
	}
}

func (c ListBigkeyScanTasksRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListBigkeyScanTasksRequestStatus) UnmarshalJSON(b []byte) error {
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
