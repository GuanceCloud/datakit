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

// 运行时类型。
type RuntimeType struct {
	value string
}

type RuntimeTypeEnum struct {
	TOMCAT8 RuntimeType
	JAVA8   RuntimeType
	PHP7    RuntimeType
	NODEJS8 RuntimeType
	DOCKER  RuntimeType
	PYTHON3 RuntimeType
	CUSTOM  RuntimeType
}

func GetRuntimeTypeEnum() RuntimeTypeEnum {
	return RuntimeTypeEnum{
		TOMCAT8: RuntimeType{
			value: "Tomcat8",
		},
		JAVA8: RuntimeType{
			value: "Java8",
		},
		PHP7: RuntimeType{
			value: "Php7",
		},
		NODEJS8: RuntimeType{
			value: "Nodejs8",
		},
		DOCKER: RuntimeType{
			value: "Docker",
		},
		PYTHON3: RuntimeType{
			value: "Python3",
		},
		CUSTOM: RuntimeType{
			value: "Custom",
		},
	}
}

func (c RuntimeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RuntimeType) UnmarshalJSON(b []byte) error {
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
