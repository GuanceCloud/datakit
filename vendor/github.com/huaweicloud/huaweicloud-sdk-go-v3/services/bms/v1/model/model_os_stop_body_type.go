/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 数据结构说明
type OsStopBodyType struct {
	// 关机类型：SOFT：普通关机（默认）。HARD：强制关机。
	Type OsStopBodyTypeType `json:"type"`
	// 裸金属服务器ID列表，详情请参见表3 servers字段数据结构说明。
	Servers []ServersList `json:"servers"`
}

func (o OsStopBodyType) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OsStopBodyType struct{}"
	}

	return strings.Join([]string{"OsStopBodyType", string(data)}, " ")
}

type OsStopBodyTypeType struct {
	value string
}

type OsStopBodyTypeTypeEnum struct {
	SOFT OsStopBodyTypeType
	HARD OsStopBodyTypeType
}

func GetOsStopBodyTypeTypeEnum() OsStopBodyTypeTypeEnum {
	return OsStopBodyTypeTypeEnum{
		SOFT: OsStopBodyTypeType{
			value: "SOFT",
		},
		HARD: OsStopBodyTypeType{
			value: "HARD",
		},
	}
}

func (c OsStopBodyTypeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OsStopBodyTypeType) UnmarshalJSON(b []byte) error {
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
