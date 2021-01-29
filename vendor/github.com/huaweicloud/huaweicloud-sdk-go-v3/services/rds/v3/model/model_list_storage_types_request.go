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
type ListStorageTypesRequest struct {
	XLanguage    *string                             `json:"X-Language,omitempty"`
	DatabaseName ListStorageTypesRequestDatabaseName `json:"database_name"`
	VersionName  string                              `json:"version_name"`
	HaMode       *ListStorageTypesRequestHaMode      `json:"ha_mode,omitempty"`
}

func (o ListStorageTypesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStorageTypesRequest struct{}"
	}

	return strings.Join([]string{"ListStorageTypesRequest", string(data)}, " ")
}

type ListStorageTypesRequestDatabaseName struct {
	value string
}

type ListStorageTypesRequestDatabaseNameEnum struct {
	MY_SQL      ListStorageTypesRequestDatabaseName
	POSTGRE_SQL ListStorageTypesRequestDatabaseName
	SQL_SERVER  ListStorageTypesRequestDatabaseName
}

func GetListStorageTypesRequestDatabaseNameEnum() ListStorageTypesRequestDatabaseNameEnum {
	return ListStorageTypesRequestDatabaseNameEnum{
		MY_SQL: ListStorageTypesRequestDatabaseName{
			value: "MySQL",
		},
		POSTGRE_SQL: ListStorageTypesRequestDatabaseName{
			value: "PostgreSQL",
		},
		SQL_SERVER: ListStorageTypesRequestDatabaseName{
			value: "SQLServer",
		},
	}
}

func (c ListStorageTypesRequestDatabaseName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListStorageTypesRequestDatabaseName) UnmarshalJSON(b []byte) error {
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

type ListStorageTypesRequestHaMode struct {
	value string
}

type ListStorageTypesRequestHaModeEnum struct {
	HA      ListStorageTypesRequestHaMode
	SINGLE  ListStorageTypesRequestHaMode
	REPLICA ListStorageTypesRequestHaMode
}

func GetListStorageTypesRequestHaModeEnum() ListStorageTypesRequestHaModeEnum {
	return ListStorageTypesRequestHaModeEnum{
		HA: ListStorageTypesRequestHaMode{
			value: "ha",
		},
		SINGLE: ListStorageTypesRequestHaMode{
			value: "single",
		},
		REPLICA: ListStorageTypesRequestHaMode{
			value: "replica",
		},
	}
}

func (c ListStorageTypesRequestHaMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListStorageTypesRequestHaMode) UnmarshalJSON(b []byte) error {
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
