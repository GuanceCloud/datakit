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
type ListStorageTypeRequest struct {
	EngineName *ListStorageTypeRequestEngineName `json:"engine_name,omitempty"`
}

func (o ListStorageTypeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStorageTypeRequest struct{}"
	}

	return strings.Join([]string{"ListStorageTypeRequest", string(data)}, " ")
}

type ListStorageTypeRequestEngineName struct {
	value string
}

type ListStorageTypeRequestEngineNameEnum struct {
	DDS_COMMUNITY ListStorageTypeRequestEngineName
}

func GetListStorageTypeRequestEngineNameEnum() ListStorageTypeRequestEngineNameEnum {
	return ListStorageTypeRequestEngineNameEnum{
		DDS_COMMUNITY: ListStorageTypeRequestEngineName{
			value: "DDS-Community",
		},
	}
}

func (c ListStorageTypeRequestEngineName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListStorageTypeRequestEngineName) UnmarshalJSON(b []byte) error {
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
