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
type CreateConfigurationRequest struct {
	XLanguage *CreateConfigurationRequestXLanguage `json:"X-Language,omitempty"`
	Body      *ConfigurationForCreation            `json:"body,omitempty"`
}

func (o CreateConfigurationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateConfigurationRequest struct{}"
	}

	return strings.Join([]string{"CreateConfigurationRequest", string(data)}, " ")
}

type CreateConfigurationRequestXLanguage struct {
	value string
}

type CreateConfigurationRequestXLanguageEnum struct {
	ZH_CN CreateConfigurationRequestXLanguage
	EN_US CreateConfigurationRequestXLanguage
}

func GetCreateConfigurationRequestXLanguageEnum() CreateConfigurationRequestXLanguageEnum {
	return CreateConfigurationRequestXLanguageEnum{
		ZH_CN: CreateConfigurationRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: CreateConfigurationRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c CreateConfigurationRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateConfigurationRequestXLanguage) UnmarshalJSON(b []byte) error {
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
