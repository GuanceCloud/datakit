/*
 * CCE
 *
 * CCE开放API
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
type ListClustersRequest struct {
	ContentType string                     `json:"Content-Type"`
	ErrorStatus *string                    `json:"errorStatus,omitempty"`
	Detail      *string                    `json:"detail,omitempty"`
	Status      *ListClustersRequestStatus `json:"status,omitempty"`
	Type        *ListClustersRequestType   `json:"type,omitempty"`
	Version     *string                    `json:"version,omitempty"`
}

func (o ListClustersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListClustersRequest struct{}"
	}

	return strings.Join([]string{"ListClustersRequest", string(data)}, " ")
}

type ListClustersRequestStatus struct {
	value string
}

type ListClustersRequestStatusEnum struct {
	AVAILABLE    ListClustersRequestStatus
	UNAVAILABLE  ListClustersRequestStatus
	SCALING_UP   ListClustersRequestStatus
	SCALING_DOWN ListClustersRequestStatus
	CREATING     ListClustersRequestStatus
	DELETING     ListClustersRequestStatus
	UPGRADING    ListClustersRequestStatus
	RESIZING     ListClustersRequestStatus
	EMPTY        ListClustersRequestStatus
}

func GetListClustersRequestStatusEnum() ListClustersRequestStatusEnum {
	return ListClustersRequestStatusEnum{
		AVAILABLE: ListClustersRequestStatus{
			value: "Available",
		},
		UNAVAILABLE: ListClustersRequestStatus{
			value: "Unavailable",
		},
		SCALING_UP: ListClustersRequestStatus{
			value: "ScalingUp",
		},
		SCALING_DOWN: ListClustersRequestStatus{
			value: "ScalingDown",
		},
		CREATING: ListClustersRequestStatus{
			value: "Creating",
		},
		DELETING: ListClustersRequestStatus{
			value: "Deleting",
		},
		UPGRADING: ListClustersRequestStatus{
			value: "Upgrading",
		},
		RESIZING: ListClustersRequestStatus{
			value: "Resizing",
		},
		EMPTY: ListClustersRequestStatus{
			value: "Empty",
		},
	}
}

func (c ListClustersRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListClustersRequestStatus) UnmarshalJSON(b []byte) error {
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

type ListClustersRequestType struct {
	value string
}

type ListClustersRequestTypeEnum struct {
	VIRTUAL_MACHINE ListClustersRequestType
	BARE_METAL      ListClustersRequestType
	ARM64           ListClustersRequestType
}

func GetListClustersRequestTypeEnum() ListClustersRequestTypeEnum {
	return ListClustersRequestTypeEnum{
		VIRTUAL_MACHINE: ListClustersRequestType{
			value: "VirtualMachine",
		},
		BARE_METAL: ListClustersRequestType{
			value: "BareMetal",
		},
		ARM64: ListClustersRequestType{
			value: "ARM64",
		},
	}
}

func (c ListClustersRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListClustersRequestType) UnmarshalJSON(b []byte) error {
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
