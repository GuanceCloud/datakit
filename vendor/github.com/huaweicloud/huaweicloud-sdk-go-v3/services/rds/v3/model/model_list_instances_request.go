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

// Request Object
type ListInstancesRequest struct {
	ContentType   *string                            `json:"Content-Type,omitempty"`
	XLanguage     *ListInstancesRequestXLanguage     `json:"X-Language,omitempty"`
	Id            *string                            `json:"id,omitempty"`
	Name          *string                            `json:"name,omitempty"`
	Type          *ListInstancesRequestType          `json:"type,omitempty"`
	DatastoreType *ListInstancesRequestDatastoreType `json:"datastore_type,omitempty"`
	VpcId         *string                            `json:"vpc_id,omitempty"`
	SubnetId      *string                            `json:"subnet_id,omitempty"`
	Offset        *int32                             `json:"offset,omitempty"`
	Limit         *int32                             `json:"limit,omitempty"`
	Tags          *string                            `json:"tags,omitempty"`
}

func (o ListInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListInstancesRequest", string(data)}, " ")
}

type ListInstancesRequestXLanguage struct {
	value string
}

type ListInstancesRequestXLanguageEnum struct {
	ZH_CN ListInstancesRequestXLanguage
	EN_US ListInstancesRequestXLanguage
}

func GetListInstancesRequestXLanguageEnum() ListInstancesRequestXLanguageEnum {
	return ListInstancesRequestXLanguageEnum{
		ZH_CN: ListInstancesRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListInstancesRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListInstancesRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesRequestXLanguage) UnmarshalJSON(b []byte) error {
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

type ListInstancesRequestType struct {
	value string
}

type ListInstancesRequestTypeEnum struct {
	SINGLE  ListInstancesRequestType
	HA      ListInstancesRequestType
	REPLICA ListInstancesRequestType
}

func GetListInstancesRequestTypeEnum() ListInstancesRequestTypeEnum {
	return ListInstancesRequestTypeEnum{
		SINGLE: ListInstancesRequestType{
			value: "Single",
		},
		HA: ListInstancesRequestType{
			value: "Ha",
		},
		REPLICA: ListInstancesRequestType{
			value: "Replica",
		},
	}
}

func (c ListInstancesRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesRequestType) UnmarshalJSON(b []byte) error {
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
	MY_SQL      ListInstancesRequestDatastoreType
	POSTGRE_SQL ListInstancesRequestDatastoreType
	SQL_SERVER  ListInstancesRequestDatastoreType
}

func GetListInstancesRequestDatastoreTypeEnum() ListInstancesRequestDatastoreTypeEnum {
	return ListInstancesRequestDatastoreTypeEnum{
		MY_SQL: ListInstancesRequestDatastoreType{
			value: "MySQL",
		},
		POSTGRE_SQL: ListInstancesRequestDatastoreType{
			value: "PostgreSQL",
		},
		SQL_SERVER: ListInstancesRequestDatastoreType{
			value: "SQLServer",
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
