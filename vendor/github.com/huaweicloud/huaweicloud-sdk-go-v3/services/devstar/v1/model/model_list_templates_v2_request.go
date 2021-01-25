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
type ListTemplatesV2Request struct {
	XLanguage *ListTemplatesV2RequestXLanguage `json:"X-Language,omitempty"`
	ActionId  string                           `json:"action_id"`
	Body      *TemplateQueryV2                 `json:"body,omitempty"`
}

func (o ListTemplatesV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTemplatesV2Request struct{}"
	}

	return strings.Join([]string{"ListTemplatesV2Request", string(data)}, " ")
}

type ListTemplatesV2RequestXLanguage struct {
	value string
}

type ListTemplatesV2RequestXLanguageEnum struct {
	ZH_CN ListTemplatesV2RequestXLanguage
	EN_US ListTemplatesV2RequestXLanguage
}

func GetListTemplatesV2RequestXLanguageEnum() ListTemplatesV2RequestXLanguageEnum {
	return ListTemplatesV2RequestXLanguageEnum{
		ZH_CN: ListTemplatesV2RequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListTemplatesV2RequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListTemplatesV2RequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListTemplatesV2RequestXLanguage) UnmarshalJSON(b []byte) error {
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
