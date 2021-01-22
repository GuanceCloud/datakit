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
type MigrateFollowerRequest struct {
	XLanguage  *MigrateFollowerRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                           `json:"instance_id"`
	Body       *FollowerMigrateRequest          `json:"body,omitempty"`
}

func (o MigrateFollowerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MigrateFollowerRequest struct{}"
	}

	return strings.Join([]string{"MigrateFollowerRequest", string(data)}, " ")
}

type MigrateFollowerRequestXLanguage struct {
	value string
}

type MigrateFollowerRequestXLanguageEnum struct {
	ZH_CN MigrateFollowerRequestXLanguage
	EN_US MigrateFollowerRequestXLanguage
}

func GetMigrateFollowerRequestXLanguageEnum() MigrateFollowerRequestXLanguageEnum {
	return MigrateFollowerRequestXLanguageEnum{
		ZH_CN: MigrateFollowerRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: MigrateFollowerRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c MigrateFollowerRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *MigrateFollowerRequestXLanguage) UnmarshalJSON(b []byte) error {
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
