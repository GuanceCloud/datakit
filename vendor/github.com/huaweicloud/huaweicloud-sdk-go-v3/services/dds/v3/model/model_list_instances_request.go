/*
 * DDS
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

// Request Object
type ListInstancesRequest struct {
	Id            *string                            `json:"id,omitempty"`
	Name          *string                            `json:"name,omitempty"`
	Mode          *ListInstancesRequestMode          `json:"mode,omitempty"`
	DatastoreType *ListInstancesRequestDatastoreType `json:"datastore_type,omitempty"`
	VpcId         *string                            `json:"vpc_id,omitempty"`
	SubnetId      *string                            `json:"subnet_id,omitempty"`
	Offset        *int32                             `json:"offset,omitempty"`
	Limit         *int32                             `json:"limit,omitempty"`
}

func (o ListInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListInstancesRequest", string(data)}, " ")
}

type ListInstancesRequestMode struct {
	value string
}

type ListInstancesRequestModeEnum struct {
	SHARDING    ListInstancesRequestMode
	REPLICA_SET ListInstancesRequestMode
	SINGLE      ListInstancesRequestMode
}

func GetListInstancesRequestModeEnum() ListInstancesRequestModeEnum {
	return ListInstancesRequestModeEnum{
		SHARDING: ListInstancesRequestMode{
			value: "Sharding",
		},
		REPLICA_SET: ListInstancesRequestMode{
			value: "ReplicaSet",
		},
		SINGLE: ListInstancesRequestMode{
			value: "Single",
		},
	}
}

func (c ListInstancesRequestMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesRequestMode) UnmarshalJSON(b []byte) error {
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

type ListInstancesRequestDatastoreType struct {
	value string
}

type ListInstancesRequestDatastoreTypeEnum struct {
	DDS_COMMUNITY ListInstancesRequestDatastoreType
}

func GetListInstancesRequestDatastoreTypeEnum() ListInstancesRequestDatastoreTypeEnum {
	return ListInstancesRequestDatastoreTypeEnum{
		DDS_COMMUNITY: ListInstancesRequestDatastoreType{
			value: "DDS-Community",
		},
	}
}

func (c ListInstancesRequestDatastoreType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesRequestDatastoreType) UnmarshalJSON(b []byte) error {
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
