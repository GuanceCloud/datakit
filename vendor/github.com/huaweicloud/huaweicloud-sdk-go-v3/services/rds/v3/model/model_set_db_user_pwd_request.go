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
type SetDbUserPwdRequest struct {
	XLanguage  *SetDbUserPwdRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                        `json:"instance_id"`
	Body       *DbUserPwdRequest             `json:"body,omitempty"`
}

func (o SetDbUserPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetDbUserPwdRequest struct{}"
	}

	return strings.Join([]string{"SetDbUserPwdRequest", string(data)}, " ")
}

type SetDbUserPwdRequestXLanguage struct {
	value string
}

type SetDbUserPwdRequestXLanguageEnum struct {
	ZH_CN SetDbUserPwdRequestXLanguage
	EN_US SetDbUserPwdRequestXLanguage
}

func GetSetDbUserPwdRequestXLanguageEnum() SetDbUserPwdRequestXLanguageEnum {
	return SetDbUserPwdRequestXLanguageEnum{
		ZH_CN: SetDbUserPwdRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: SetDbUserPwdRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c SetDbUserPwdRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SetDbUserPwdRequestXLanguage) UnmarshalJSON(b []byte) error {
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
