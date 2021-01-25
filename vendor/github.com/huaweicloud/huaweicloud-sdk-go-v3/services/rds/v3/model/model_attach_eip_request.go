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
type AttachEipRequest struct {
	XLanguage  *AttachEipRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                     `json:"instance_id"`
	Body       *BindEipRequest            `json:"body,omitempty"`
}

func (o AttachEipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachEipRequest struct{}"
	}

	return strings.Join([]string{"AttachEipRequest", string(data)}, " ")
}

type AttachEipRequestXLanguage struct {
	value string
}

type AttachEipRequestXLanguageEnum struct {
	ZH_CN AttachEipRequestXLanguage
	EN_US AttachEipRequestXLanguage
}

func GetAttachEipRequestXLanguageEnum() AttachEipRequestXLanguageEnum {
	return AttachEipRequestXLanguageEnum{
		ZH_CN: AttachEipRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: AttachEipRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c AttachEipRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AttachEipRequestXLanguage) UnmarshalJSON(b []byte) error {
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
