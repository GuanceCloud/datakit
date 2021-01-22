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
type ListFlavorsRequest struct {
	Region     string                        `json:"region"`
	EngineName *ListFlavorsRequestEngineName `json:"engine_name,omitempty"`
}

func (o ListFlavorsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFlavorsRequest struct{}"
	}

	return strings.Join([]string{"ListFlavorsRequest", string(data)}, " ")
}

type ListFlavorsRequestEngineName struct {
	value string
}

type ListFlavorsRequestEngineNameEnum struct {
	DDS_COMMUNITY ListFlavorsRequestEngineName
}

func GetListFlavorsRequestEngineNameEnum() ListFlavorsRequestEngineNameEnum {
	return ListFlavorsRequestEngineNameEnum{
		DDS_COMMUNITY: ListFlavorsRequestEngineName{
			value: "DDS-Community",
		},
	}
}

func (c ListFlavorsRequestEngineName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListFlavorsRequestEngineName) UnmarshalJSON(b []byte) error {
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
