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
type StartInstanceActionRequest struct {
	XLanguage  *StartInstanceActionRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                               `json:"instance_id"`
	Body       *InstanceActionRequest               `json:"body,omitempty"`
}

func (o StartInstanceActionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartInstanceActionRequest struct{}"
	}

	return strings.Join([]string{"StartInstanceActionRequest", string(data)}, " ")
}

type StartInstanceActionRequestXLanguage struct {
	value string
}

type StartInstanceActionRequestXLanguageEnum struct {
	ZH_CN StartInstanceActionRequestXLanguage
	EN_US StartInstanceActionRequestXLanguage
}

func GetStartInstanceActionRequestXLanguageEnum() StartInstanceActionRequestXLanguageEnum {
	return StartInstanceActionRequestXLanguageEnum{
		ZH_CN: StartInstanceActionRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: StartInstanceActionRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c StartInstanceActionRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StartInstanceActionRequestXLanguage) UnmarshalJSON(b []byte) error {
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
