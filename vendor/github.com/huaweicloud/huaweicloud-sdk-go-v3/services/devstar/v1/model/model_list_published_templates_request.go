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
type ListPublishedTemplatesRequest struct {
	XLanguage *ListPublishedTemplatesRequestXLanguage `json:"X-Language,omitempty"`
	Keyword   *string                                 `json:"keyword,omitempty"`
	Offset    *int32                                  `json:"offset,omitempty"`
	Limit     *int32                                  `json:"limit,omitempty"`
}

func (o ListPublishedTemplatesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublishedTemplatesRequest struct{}"
	}

	return strings.Join([]string{"ListPublishedTemplatesRequest", string(data)}, " ")
}

type ListPublishedTemplatesRequestXLanguage struct {
	value string
}

type ListPublishedTemplatesRequestXLanguageEnum struct {
	ZH_CN ListPublishedTemplatesRequestXLanguage
	EN_US ListPublishedTemplatesRequestXLanguage
}

func GetListPublishedTemplatesRequestXLanguageEnum() ListPublishedTemplatesRequestXLanguageEnum {
	return ListPublishedTemplatesRequestXLanguageEnum{
		ZH_CN: ListPublishedTemplatesRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListPublishedTemplatesRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListPublishedTemplatesRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublishedTemplatesRequestXLanguage) UnmarshalJSON(b []byte) error {
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
