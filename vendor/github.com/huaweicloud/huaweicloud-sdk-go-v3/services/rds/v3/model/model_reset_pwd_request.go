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
type ResetPwdRequest struct {
	XLanguage  *ResetPwdRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                    `json:"instance_id"`
	Body       *PwdResetRequest          `json:"body,omitempty"`
}

func (o ResetPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPwdRequest struct{}"
	}

	return strings.Join([]string{"ResetPwdRequest", string(data)}, " ")
}

type ResetPwdRequestXLanguage struct {
	value string
}

type ResetPwdRequestXLanguageEnum struct {
	ZH_CN ResetPwdRequestXLanguage
	EN_US ResetPwdRequestXLanguage
}

func GetResetPwdRequestXLanguageEnum() ResetPwdRequestXLanguageEnum {
	return ResetPwdRequestXLanguageEnum{
		ZH_CN: ResetPwdRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ResetPwdRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ResetPwdRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResetPwdRequestXLanguage) UnmarshalJSON(b []byte) error {
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
