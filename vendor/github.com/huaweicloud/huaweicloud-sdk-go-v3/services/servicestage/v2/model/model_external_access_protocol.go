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

// 协议。
type ExternalAccessProtocol struct {
	value string
}

type ExternalAccessProtocolEnum struct {
	HTTP  ExternalAccessProtocol
	HTTPS ExternalAccessProtocol
}

func GetExternalAccessProtocolEnum() ExternalAccessProtocolEnum {
	return ExternalAccessProtocolEnum{
		HTTP: ExternalAccessProtocol{
			value: "HTTP",
		},
		HTTPS: ExternalAccessProtocol{
			value: "HTTPS",
		},
	}
}

func (c ExternalAccessProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExternalAccessProtocol) UnmarshalJSON(b []byte) error {
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
