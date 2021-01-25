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

// 来源类型。支持源码code和artifact软件包。
type SourceKind struct {
	value string
}

type SourceKindEnum struct {
	CODE     SourceKind
	ARTIFACT SourceKind
}

func GetSourceKindEnum() SourceKindEnum {
	return SourceKindEnum{
		CODE: SourceKind{
			value: "code",
		},
		ARTIFACT: SourceKind{
			value: "artifact",
		},
	}
}

func (c SourceKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SourceKind) UnmarshalJSON(b []byte) error {
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
