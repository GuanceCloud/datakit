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
type ChangeFailoverModeRequest struct {
	XLanguage  *ChangeFailoverModeRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                              `json:"instance_id"`
	Body       *FailoverModeRequest                `json:"body,omitempty"`
}

func (o ChangeFailoverModeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeFailoverModeRequest struct{}"
	}

	return strings.Join([]string{"ChangeFailoverModeRequest", string(data)}, " ")
}

type ChangeFailoverModeRequestXLanguage struct {
	value string
}

type ChangeFailoverModeRequestXLanguageEnum struct {
	ZH_CN ChangeFailoverModeRequestXLanguage
	EN_US ChangeFailoverModeRequestXLanguage
}

func GetChangeFailoverModeRequestXLanguageEnum() ChangeFailoverModeRequestXLanguageEnum {
	return ChangeFailoverModeRequestXLanguageEnum{
		ZH_CN: ChangeFailoverModeRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ChangeFailoverModeRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ChangeFailoverModeRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeFailoverModeRequestXLanguage) UnmarshalJSON(b []byte) error {
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
