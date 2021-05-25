/*
 * RMS
 *
 * Resource Manager Api
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
type ListProvidersRequest struct {
	Offset    *int32                         `json:"offset,omitempty"`
	Limit     *int32                         `json:"limit,omitempty"`
	XLanguage *ListProvidersRequestXLanguage `json:"X-Language,omitempty"`
}

func (o ListProvidersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProvidersRequest struct{}"
	}

	return strings.Join([]string{"ListProvidersRequest", string(data)}, " ")
}

type ListProvidersRequestXLanguage struct {
	value string
}

type ListProvidersRequestXLanguageEnum struct {
	ZH_CN ListProvidersRequestXLanguage
	EN_US ListProvidersRequestXLanguage
}

func GetListProvidersRequestXLanguageEnum() ListProvidersRequestXLanguageEnum {
	return ListProvidersRequestXLanguageEnum{
		ZH_CN: ListProvidersRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListProvidersRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListProvidersRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListProvidersRequestXLanguage) UnmarshalJSON(b []byte) error {
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
