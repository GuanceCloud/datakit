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
type UpdatePortRequest struct {
	XLanguage  *UpdatePortRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                      `json:"instance_id"`
	Body       *UpdateDbPortRequest        `json:"body,omitempty"`
}

func (o UpdatePortRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePortRequest struct{}"
	}

	return strings.Join([]string{"UpdatePortRequest", string(data)}, " ")
}

type UpdatePortRequestXLanguage struct {
	value string
}

type UpdatePortRequestXLanguageEnum struct {
	ZH_CN UpdatePortRequestXLanguage
	EN_US UpdatePortRequestXLanguage
}

func GetUpdatePortRequestXLanguageEnum() UpdatePortRequestXLanguageEnum {
	return UpdatePortRequestXLanguageEnum{
		ZH_CN: UpdatePortRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: UpdatePortRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c UpdatePortRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdatePortRequestXLanguage) UnmarshalJSON(b []byte) error {
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
