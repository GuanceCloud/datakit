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

// 外网访问。
type ExternalAccessesCreate struct {
	// 协议，支持http、https。
	Protocol ExternalAccessesCreateProtocol `json:"protocol"`
	// 访问地址。
	Address string `json:"address"`
	// 端口号。
	ForwardPort int32 `json:"forward_port"`
}

func (o ExternalAccessesCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExternalAccessesCreate struct{}"
	}

	return strings.Join([]string{"ExternalAccessesCreate", string(data)}, " ")
}

type ExternalAccessesCreateProtocol struct {
	value string
}

type ExternalAccessesCreateProtocolEnum struct {
	HTTP  ExternalAccessesCreateProtocol
	HTTPS ExternalAccessesCreateProtocol
}

func GetExternalAccessesCreateProtocolEnum() ExternalAccessesCreateProtocolEnum {
	return ExternalAccessesCreateProtocolEnum{
		HTTP: ExternalAccessesCreateProtocol{
			value: "http",
		},
		HTTPS: ExternalAccessesCreateProtocol{
			value: "https",
		},
	}
}

func (c ExternalAccessesCreateProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExternalAccessesCreateProtocol) UnmarshalJSON(b []byte) error {
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
