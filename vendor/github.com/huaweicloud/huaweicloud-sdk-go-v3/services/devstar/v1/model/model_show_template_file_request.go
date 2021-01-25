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
type ShowTemplateFileRequest struct {
	XLanguage  *ShowTemplateFileRequestXLanguage `json:"X-Language,omitempty"`
	TemplateId string                            `json:"template_id"`
	FilePath   string                            `json:"file_path"`
	Type       *ShowTemplateFileRequestType      `json:"type,omitempty"`
}

func (o ShowTemplateFileRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateFileRequest struct{}"
	}

	return strings.Join([]string{"ShowTemplateFileRequest", string(data)}, " ")
}

type ShowTemplateFileRequestXLanguage struct {
	value string
}

type ShowTemplateFileRequestXLanguageEnum struct {
	ZH_CN ShowTemplateFileRequestXLanguage
	EN_US ShowTemplateFileRequestXLanguage
}

func GetShowTemplateFileRequestXLanguageEnum() ShowTemplateFileRequestXLanguageEnum {
	return ShowTemplateFileRequestXLanguageEnum{
		ZH_CN: ShowTemplateFileRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ShowTemplateFileRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ShowTemplateFileRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowTemplateFileRequestXLanguage) UnmarshalJSON(b []byte) error {
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

type ShowTemplateFileRequestType struct {
	value string
}

type ShowTemplateFileRequestTypeEnum struct {
	SOURCE_PACKAGE ShowTemplateFileRequestType
	INTRODUCTION   ShowTemplateFileRequestType
}

func GetShowTemplateFileRequestTypeEnum() ShowTemplateFileRequestTypeEnum {
	return ShowTemplateFileRequestTypeEnum{
		SOURCE_PACKAGE: ShowTemplateFileRequestType{
			value: "source-package",
		},
		INTRODUCTION: ShowTemplateFileRequestType{
			value: "introduction",
		},
	}
}

func (c ShowTemplateFileRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowTemplateFileRequestType) UnmarshalJSON(b []byte) error {
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
