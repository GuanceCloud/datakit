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
type ShowTemplateDetailRequest struct {
	XLanguage  *ShowTemplateDetailRequestXLanguage `json:"X-Language,omitempty"`
	TemplateId string                              `json:"template_id"`
}

func (o ShowTemplateDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowTemplateDetailRequest", string(data)}, " ")
}

type ShowTemplateDetailRequestXLanguage struct {
	value string
}

type ShowTemplateDetailRequestXLanguageEnum struct {
	ZH_CN ShowTemplateDetailRequestXLanguage
	EN_US ShowTemplateDetailRequestXLanguage
}

func GetShowTemplateDetailRequestXLanguageEnum() ShowTemplateDetailRequestXLanguageEnum {
	return ShowTemplateDetailRequestXLanguageEnum{
		ZH_CN: ShowTemplateDetailRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ShowTemplateDetailRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ShowTemplateDetailRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowTemplateDetailRequestXLanguage) UnmarshalJSON(b []byte) error {
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
