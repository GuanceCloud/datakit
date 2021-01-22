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
type ServersInfoType struct {
	// 重启类型：SOFT：普通重启。HARD：强制重启。
	Type ServersInfoTypeType `json:"type"`
	// 裸金属服务器ID列表，详情请参见表3 servers字段数据结构说明。
	Servers []ServersList `json:"servers"`
}

func (o ServersInfoType) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServersInfoType struct{}"
	}

	return strings.Join([]string{"ServersInfoType", string(data)}, " ")
}

type ServersInfoTypeType struct {
	value string
}

type ServersInfoTypeTypeEnum struct {
	SOFT ServersInfoTypeType
	HARD ServersInfoTypeType
}

func GetServersInfoTypeTypeEnum() ServersInfoTypeTypeEnum {
	return ServersInfoTypeTypeEnum{
		SOFT: ServersInfoTypeType{
			value: "SOFT",
		},
		HARD: ServersInfoTypeType{
			value: "HARD",
		},
	}
}

func (c ServersInfoTypeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServersInfoTypeType) UnmarshalJSON(b []byte) error {
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
