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
type SetSecurityGroupRequest struct {
	XLanguage  *SetSecurityGroupRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                            `json:"instance_id"`
	Body       *SecurityGroupRequest             `json:"body,omitempty"`
}

func (o SetSecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"SetSecurityGroupRequest", string(data)}, " ")
}

type SetSecurityGroupRequestXLanguage struct {
	value string
}

type SetSecurityGroupRequestXLanguageEnum struct {
	ZH_CN SetSecurityGroupRequestXLanguage
	EN_US SetSecurityGroupRequestXLanguage
}

func GetSetSecurityGroupRequestXLanguageEnum() SetSecurityGroupRequestXLanguageEnum {
	return SetSecurityGroupRequestXLanguageEnum{
		ZH_CN: SetSecurityGroupRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: SetSecurityGroupRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c SetSecurityGroupRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SetSecurityGroupRequestXLanguage) UnmarshalJSON(b []byte) error {
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
