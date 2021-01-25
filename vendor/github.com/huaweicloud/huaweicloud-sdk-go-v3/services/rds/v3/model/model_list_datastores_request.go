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
type ListDatastoresRequest struct {
	XLanguage    *string                           `json:"X-Language,omitempty"`
	DatabaseName ListDatastoresRequestDatabaseName `json:"database_name"`
}

func (o ListDatastoresRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDatastoresRequest struct{}"
	}

	return strings.Join([]string{"ListDatastoresRequest", string(data)}, " ")
}

type ListDatastoresRequestDatabaseName struct {
	value string
}

type ListDatastoresRequestDatabaseNameEnum struct {
	MY_SQL      ListDatastoresRequestDatabaseName
	POSTGRE_SQL ListDatastoresRequestDatabaseName
	SQL_SERVER  ListDatastoresRequestDatabaseName
}

func GetListDatastoresRequestDatabaseNameEnum() ListDatastoresRequestDatabaseNameEnum {
	return ListDatastoresRequestDatabaseNameEnum{
		MY_SQL: ListDatastoresRequestDatabaseName{
			value: "MySQL",
		},
		POSTGRE_SQL: ListDatastoresRequestDatabaseName{
			value: "PostgreSQL",
		},
		SQL_SERVER: ListDatastoresRequestDatabaseName{
			value: "SQLServer",
		},
	}
}

func (c ListDatastoresRequestDatabaseName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListDatastoresRequestDatabaseName) UnmarshalJSON(b []byte) error {
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
