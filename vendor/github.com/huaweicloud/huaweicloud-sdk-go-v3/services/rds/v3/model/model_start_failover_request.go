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
type StartFailoverRequest struct {
	XLanguage  *StartFailoverRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                         `json:"instance_id"`
}

func (o StartFailoverRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartFailoverRequest struct{}"
	}

	return strings.Join([]string{"StartFailoverRequest", string(data)}, " ")
}

type StartFailoverRequestXLanguage struct {
	value string
}

type StartFailoverRequestXLanguageEnum struct {
	ZH_CN StartFailoverRequestXLanguage
	EN_US StartFailoverRequestXLanguage
}

func GetStartFailoverRequestXLanguageEnum() StartFailoverRequestXLanguageEnum {
	return StartFailoverRequestXLanguageEnum{
		ZH_CN: StartFailoverRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: StartFailoverRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c StartFailoverRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StartFailoverRequestXLanguage) UnmarshalJSON(b []byte) error {
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
