/*
 * DevStar
 *
 * DevStar API
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
type ShowTemplateV3Request struct {
	XLanguage  *ShowTemplateV3RequestXLanguage `json:"X-Language,omitempty"`
	TemplateId string                          `json:"template_id"`
}

func (o ShowTemplateV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateV3Request struct{}"
	}

	return strings.Join([]string{"ShowTemplateV3Request", string(data)}, " ")
}

type ShowTemplateV3RequestXLanguage struct {
	value string
}

type ShowTemplateV3RequestXLanguageEnum struct {
	ZH_CN ShowTemplateV3RequestXLanguage
	EN_US ShowTemplateV3RequestXLanguage
}

func GetShowTemplateV3RequestXLanguageEnum() ShowTemplateV3RequestXLanguageEnum {
	return ShowTemplateV3RequestXLanguageEnum{
		ZH_CN: ShowTemplateV3RequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ShowTemplateV3RequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ShowTemplateV3RequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowTemplateV3RequestXLanguage) UnmarshalJSON(b []byte) error {
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
