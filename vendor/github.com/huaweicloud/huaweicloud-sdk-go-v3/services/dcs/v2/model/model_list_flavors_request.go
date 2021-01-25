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
type ListFlavorsRequest struct {
	SpecCode      *string                    `json:"spec_code,omitempty"`
	CacheMode     *string                    `json:"cache_mode,omitempty"`
	Engine        *string                    `json:"engine,omitempty"`
	EngineVersion *string                    `json:"engine_version,omitempty"`
	CpuType       *ListFlavorsRequestCpuType `json:"cpu_type,omitempty"`
	Capacity      *string                    `json:"capacity,omitempty"`
}

func (o ListFlavorsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFlavorsRequest struct{}"
	}

	return strings.Join([]string{"ListFlavorsRequest", string(data)}, " ")
}

type ListFlavorsRequestCpuType struct {
	value string
}

type ListFlavorsRequestCpuTypeEnum struct {
	X86_64  ListFlavorsRequestCpuType
	AARCH64 ListFlavorsRequestCpuType
}

func GetListFlavorsRequestCpuTypeEnum() ListFlavorsRequestCpuTypeEnum {
	return ListFlavorsRequestCpuTypeEnum{
		X86_64: ListFlavorsRequestCpuType{
			value: "x86_64",
		},
		AARCH64: ListFlavorsRequestCpuType{
			value: "aarch64",
		},
	}
}

func (c ListFlavorsRequestCpuType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListFlavorsRequestCpuType) UnmarshalJSON(b []byte) error {
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
