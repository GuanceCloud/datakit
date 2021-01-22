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
type DownloadApplicationCodeRequest struct {
	XLanguage *DownloadApplicationCodeRequestXLanguage `json:"X-Language,omitempty"`
	JobId     string                                   `json:"job_id"`
}

func (o DownloadApplicationCodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadApplicationCodeRequest struct{}"
	}

	return strings.Join([]string{"DownloadApplicationCodeRequest", string(data)}, " ")
}

type DownloadApplicationCodeRequestXLanguage struct {
	value string
}

type DownloadApplicationCodeRequestXLanguageEnum struct {
	ZH_CN DownloadApplicationCodeRequestXLanguage
	EN_US DownloadApplicationCodeRequestXLanguage
}

func GetDownloadApplicationCodeRequestXLanguageEnum() DownloadApplicationCodeRequestXLanguageEnum {
	return DownloadApplicationCodeRequestXLanguageEnum{
		ZH_CN: DownloadApplicationCodeRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: DownloadApplicationCodeRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c DownloadApplicationCodeRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DownloadApplicationCodeRequestXLanguage) UnmarshalJSON(b []byte) error {
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
