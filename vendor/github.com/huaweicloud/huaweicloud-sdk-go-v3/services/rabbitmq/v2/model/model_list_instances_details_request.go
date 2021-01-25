/*
 * RabbitMQ
 *
 * RabbitMQ Document API
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
type ListInstancesDetailsRequest struct {
	ProjectId           string                                     `json:"project_id"`
	Engine              ListInstancesDetailsRequestEngine          `json:"engine"`
	Name                *string                                    `json:"name,omitempty"`
	InstanceId          *string                                    `json:"instance_id,omitempty"`
	Status              *ListInstancesDetailsRequestStatus         `json:"status,omitempty"`
	IncludeFailure      *ListInstancesDetailsRequestIncludeFailure `json:"include_failure,omitempty"`
	ExactMatchName      *ListInstancesDetailsRequestExactMatchName `json:"exact_match_name,omitempty"`
	EnterpriseProjectId *string                                    `json:"enterprise_project_id,omitempty"`
}

func (o ListInstancesDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesDetailsRequest struct{}"
	}

	return strings.Join([]string{"ListInstancesDetailsRequest", string(data)}, " ")
}

type ListInstancesDetailsRequestEngine struct {
	value string
}

type ListInstancesDetailsRequestEngineEnum struct {
	RABBITMQ ListInstancesDetailsRequestEngine
}

func GetListInstancesDetailsRequestEngineEnum() ListInstancesDetailsRequestEngineEnum {
	return ListInstancesDetailsRequestEngineEnum{
		RABBITMQ: ListInstancesDetailsRequestEngine{
			value: "rabbitmq",
		},
	}
}

func (c ListInstancesDetailsRequestEngine) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesDetailsRequestEngine) UnmarshalJSON(b []byte) error {
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

type ListInstancesDetailsRequestStatus struct {
	value string
}

type ListInstancesDetailsRequestStatusEnum struct {
	CREATING     ListInstancesDetailsRequestStatus
	CREATEFAILED ListInstancesDetailsRequestStatus
	RUNNING      ListInstancesDetailsRequestStatus
	ERROR        ListInstancesDetailsRequestStatus
	STARTING     ListInstancesDetailsRequestStatus
	RESTARTING   ListInstancesDetailsRequestStatus
	CLOSING      ListInstancesDetailsRequestStatus
	FROZEN       ListInstancesDetailsRequestStatus
}

func GetListInstancesDetailsRequestStatusEnum() ListInstancesDetailsRequestStatusEnum {
	return ListInstancesDetailsRequestStatusEnum{
		CREATING: ListInstancesDetailsRequestStatus{
			value: "CREATING",
		},
		CREATEFAILED: ListInstancesDetailsRequestStatus{
			value: "CREATEFAILED",
		},
		RUNNING: ListInstancesDetailsRequestStatus{
			value: "RUNNING",
		},
		ERROR: ListInstancesDetailsRequestStatus{
			value: "ERROR",
		},
		STARTING: ListInstancesDetailsRequestStatus{
			value: "STARTING",
		},
		RESTARTING: ListInstancesDetailsRequestStatus{
			value: "RESTARTING",
		},
		CLOSING: ListInstancesDetailsRequestStatus{
			value: "CLOSING",
		},
		FROZEN: ListInstancesDetailsRequestStatus{
			value: "FROZEN",
		},
	}
}

func (c ListInstancesDetailsRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesDetailsRequestStatus) UnmarshalJSON(b []byte) error {
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

type ListInstancesDetailsRequestIncludeFailure struct {
	value string
}

type ListInstancesDetailsRequestIncludeFailureEnum struct {
	TRUE  ListInstancesDetailsRequestIncludeFailure
	FALSE ListInstancesDetailsRequestIncludeFailure
}

func GetListInstancesDetailsRequestIncludeFailureEnum() ListInstancesDetailsRequestIncludeFailureEnum {
	return ListInstancesDetailsRequestIncludeFailureEnum{
		TRUE: ListInstancesDetailsRequestIncludeFailure{
			value: "true",
		},
		FALSE: ListInstancesDetailsRequestIncludeFailure{
			value: "false",
		},
	}
}

func (c ListInstancesDetailsRequestIncludeFailure) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesDetailsRequestIncludeFailure) UnmarshalJSON(b []byte) error {
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

type ListInstancesDetailsRequestExactMatchName struct {
	value string
}

type ListInstancesDetailsRequestExactMatchNameEnum struct {
	TRUE  ListInstancesDetailsRequestExactMatchName
	FALSE ListInstancesDetailsRequestExactMatchName
}

func GetListInstancesDetailsRequestExactMatchNameEnum() ListInstancesDetailsRequestExactMatchNameEnum {
	return ListInstancesDetailsRequestExactMatchNameEnum{
		TRUE: ListInstancesDetailsRequestExactMatchName{
			value: "true",
		},
		FALSE: ListInstancesDetailsRequestExactMatchName{
			value: "false",
		},
	}
}

func (c ListInstancesDetailsRequestExactMatchName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesDetailsRequestExactMatchName) UnmarshalJSON(b []byte) error {
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
