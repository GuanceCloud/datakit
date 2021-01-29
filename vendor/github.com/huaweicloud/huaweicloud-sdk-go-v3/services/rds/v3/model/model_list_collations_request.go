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
type ListCollationsRequest struct {
	XLanguage *ListCollationsRequestXLanguage `json:"X-Language,omitempty"`
}

func (o ListCollationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCollationsRequest struct{}"
	}

	return strings.Join([]string{"ListCollationsRequest", string(data)}, " ")
}

type ListCollationsRequestXLanguage struct {
	value string
}

type ListCollationsRequestXLanguageEnum struct {
	ZH_CN ListCollationsRequestXLanguage
	EN_US ListCollationsRequestXLanguage
}

func GetListCollationsRequestXLanguageEnum() ListCollationsRequestXLanguageEnum {
	return ListCollationsRequestXLanguageEnum{
		ZH_CN: ListCollationsRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListCollationsRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListCollationsRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListCollationsRequestXLanguage) UnmarshalJSON(b []byte) error {
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
