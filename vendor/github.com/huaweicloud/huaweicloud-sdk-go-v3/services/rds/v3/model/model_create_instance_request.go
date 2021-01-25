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
type CreateInstanceRequest struct {
	XLanguage *CreateInstanceRequestXLanguage `json:"X-Language,omitempty"`
	Body      *InstanceRequest                `json:"body,omitempty"`
}

func (o CreateInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceRequest struct{}"
	}

	return strings.Join([]string{"CreateInstanceRequest", string(data)}, " ")
}

type CreateInstanceRequestXLanguage struct {
	value string
}

type CreateInstanceRequestXLanguageEnum struct {
	ZH_CN CreateInstanceRequestXLanguage
	EN_US CreateInstanceRequestXLanguage
}

func GetCreateInstanceRequestXLanguageEnum() CreateInstanceRequestXLanguageEnum {
	return CreateInstanceRequestXLanguageEnum{
		ZH_CN: CreateInstanceRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: CreateInstanceRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c CreateInstanceRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateInstanceRequestXLanguage) UnmarshalJSON(b []byte) error {
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
