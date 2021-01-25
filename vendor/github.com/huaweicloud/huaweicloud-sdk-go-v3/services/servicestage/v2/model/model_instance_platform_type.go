/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 运行平台类型。  应用可以在不同的平台上运行，可选用的平台的类型有以下几种：cce、vmapp。
type InstancePlatformType struct {
	value string
}

type InstancePlatformTypeEnum struct {
	CCE   InstancePlatformType
	VMAPP InstancePlatformType
}

func GetInstancePlatformTypeEnum() InstancePlatformTypeEnum {
	return InstancePlatformTypeEnum{
		CCE: InstancePlatformType{
			value: "cce",
		},
		VMAPP: InstancePlatformType{
			value: "vmapp",
		},
	}
}

func (c InstancePlatformType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstancePlatformType) UnmarshalJSON(b []byte) error {
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
