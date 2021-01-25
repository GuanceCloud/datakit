/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type PrincipalRule struct {
	// 通道唯一标识符。
	Principal *string `json:"principal,omitempty"`
	// 授权用户。
	PrincipalName *string `json:"principal_name,omitempty"`
	// 授权操作类型。  - putRecords：上传数据。
	ActionType *PrincipalRuleActionType `json:"action_type,omitempty"`
	// 授权影响类型。  - accept：允许该授权操作。
	Effect *PrincipalRuleEffect `json:"effect,omitempty"`
}

func (o PrincipalRule) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PrincipalRule struct{}"
	}

	return strings.Join([]string{"PrincipalRule", string(data)}, " ")
}

type PrincipalRuleActionType struct {
	value string
}

type PrincipalRuleActionTypeEnum struct {
	PUT_RECORDS PrincipalRuleActionType
}

func GetPrincipalRuleActionTypeEnum() PrincipalRuleActionTypeEnum {
	return PrincipalRuleActionTypeEnum{
		PUT_RECORDS: PrincipalRuleActionType{
			value: "putRecords",
		},
	}
}

func (c PrincipalRuleActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PrincipalRuleActionType) UnmarshalJSON(b []byte) error {
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

type PrincipalRuleEffect struct {
	value string
}

type PrincipalRuleEffectEnum struct {
	ACCEPT PrincipalRuleEffect
}

func GetPrincipalRuleEffectEnum() PrincipalRuleEffectEnum {
	return PrincipalRuleEffectEnum{
		ACCEPT: PrincipalRuleEffect{
			value: "accept",
		},
	}
}

func (c PrincipalRuleEffect) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PrincipalRuleEffect) UnmarshalJSON(b []byte) error {
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
