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
type ListFlavorsRequest struct {
	XLanguage    *string                        `json:"X-Language,omitempty"`
	DatabaseName ListFlavorsRequestDatabaseName `json:"database_name"`
	VersionName  *string                        `json:"version_name,omitempty"`
	SpecCode     *string                        `json:"spec_code,omitempty"`
}

func (o ListFlavorsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFlavorsRequest struct{}"
	}

	return strings.Join([]string{"ListFlavorsRequest", string(data)}, " ")
}

type ListFlavorsRequestDatabaseName struct {
	value string
}

type ListFlavorsRequestDatabaseNameEnum struct {
	MY_SQL      ListFlavorsRequestDatabaseName
	POSTGRE_SQL ListFlavorsRequestDatabaseName
	SQL_SERVER  ListFlavorsRequestDatabaseName
}

func GetListFlavorsRequestDatabaseNameEnum() ListFlavorsRequestDatabaseNameEnum {
	return ListFlavorsRequestDatabaseNameEnum{
		MY_SQL: ListFlavorsRequestDatabaseName{
			value: "MySQL",
		},
		POSTGRE_SQL: ListFlavorsRequestDatabaseName{
			value: "PostgreSQL",
		},
		SQL_SERVER: ListFlavorsRequestDatabaseName{
			value: "SQLServer",
		},
	}
}

func (c ListFlavorsRequestDatabaseName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListFlavorsRequestDatabaseName) UnmarshalJSON(b []byte) error {
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
