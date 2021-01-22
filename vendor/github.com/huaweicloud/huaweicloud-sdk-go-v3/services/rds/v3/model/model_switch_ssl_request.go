/*
 * RDS
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
type SwitchSslRequest struct {
	XLanguage  *SwitchSslRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                     `json:"instance_id"`
	Body       *SslOptionRequestBody      `json:"body,omitempty"`
}

func (o SwitchSslRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SwitchSslRequest struct{}"
	}

	return strings.Join([]string{"SwitchSslRequest", string(data)}, " ")
}

type SwitchSslRequestXLanguage struct {
	value string
}

type SwitchSslRequestXLanguageEnum struct {
	ZH_CN SwitchSslRequestXLanguage
	EN_US SwitchSslRequestXLanguage
}

func GetSwitchSslRequestXLanguageEnum() SwitchSslRequestXLanguageEnum {
	return SwitchSslRequestXLanguageEnum{
		ZH_CN: SwitchSslRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: SwitchSslRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c SwitchSslRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SwitchSslRequestXLanguage) UnmarshalJSON(b []byte) error {
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
