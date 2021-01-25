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
type CreateTemplateViewHistoriesRequest struct {
	XLanguage *CreateTemplateViewHistoriesRequestXLanguage `json:"X-Language,omitempty"`
	Body      *TemplatesInfo                               `json:"body,omitempty"`
}

func (o CreateTemplateViewHistoriesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTemplateViewHistoriesRequest struct{}"
	}

	return strings.Join([]string{"CreateTemplateViewHistoriesRequest", string(data)}, " ")
}

type CreateTemplateViewHistoriesRequestXLanguage struct {
	value string
}

type CreateTemplateViewHistoriesRequestXLanguageEnum struct {
	ZH_CN CreateTemplateViewHistoriesRequestXLanguage
	EN_US CreateTemplateViewHistoriesRequestXLanguage
}

func GetCreateTemplateViewHistoriesRequestXLanguageEnum() CreateTemplateViewHistoriesRequestXLanguageEnum {
	return CreateTemplateViewHistoriesRequestXLanguageEnum{
		ZH_CN: CreateTemplateViewHistoriesRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: CreateTemplateViewHistoriesRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c CreateTemplateViewHistoriesRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateTemplateViewHistoriesRequestXLanguage) UnmarshalJSON(b []byte) error {
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
