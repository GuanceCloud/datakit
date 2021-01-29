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
type DownloadSlowlogRequest struct {
	XLanguage  *DownloadSlowlogRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                           `json:"instance_id"`
	Body       *SlowlogDownloadRequest          `json:"body,omitempty"`
}

func (o DownloadSlowlogRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadSlowlogRequest struct{}"
	}

	return strings.Join([]string{"DownloadSlowlogRequest", string(data)}, " ")
}

type DownloadSlowlogRequestXLanguage struct {
	value string
}

type DownloadSlowlogRequestXLanguageEnum struct {
	ZH_CN DownloadSlowlogRequestXLanguage
	EN_US DownloadSlowlogRequestXLanguage
}

func GetDownloadSlowlogRequestXLanguageEnum() DownloadSlowlogRequestXLanguageEnum {
	return DownloadSlowlogRequestXLanguageEnum{
		ZH_CN: DownloadSlowlogRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: DownloadSlowlogRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c DownloadSlowlogRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DownloadSlowlogRequestXLanguage) UnmarshalJSON(b []byte) error {
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
