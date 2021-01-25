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
type ShowJobDetailRequest struct {
	XLanguage *ShowJobDetailRequestXLanguage `json:"X-Language,omitempty"`
	JobId     string                         `json:"job_id"`
}

func (o ShowJobDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowJobDetailRequest", string(data)}, " ")
}

type ShowJobDetailRequestXLanguage struct {
	value string
}

type ShowJobDetailRequestXLanguageEnum struct {
	ZH_CN ShowJobDetailRequestXLanguage
	EN_US ShowJobDetailRequestXLanguage
}

func GetShowJobDetailRequestXLanguageEnum() ShowJobDetailRequestXLanguageEnum {
	return ShowJobDetailRequestXLanguageEnum{
		ZH_CN: ShowJobDetailRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ShowJobDetailRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ShowJobDetailRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowJobDetailRequestXLanguage) UnmarshalJSON(b []byte) error {
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
