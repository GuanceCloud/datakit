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

type CreatePolicyRuleRequest struct {
	// 通道唯一标识符。
	StreamId string `json:"stream_id"`
	// 授权用户。  支持通配符\"*\"，表示授权所有账号，支持多账号添加，用\",\"隔开；
	PrincipalName string `json:"principal_name"`
	// 授权操作类型。  - putRecords：上传数据。 - getRecords：下载数据。
	ActionType CreatePolicyRuleRequestActionType `json:"action_type"`
	// 授权影响类型。  - accept：允许该授权操作。
	Effect CreatePolicyRuleRequestEffect `json:"effect"`
}

func (o CreatePolicyRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePolicyRuleRequest struct{}"
	}

	return strings.Join([]string{"CreatePolicyRuleRequest", string(data)}, " ")
}

type CreatePolicyRuleRequestActionType struct {
	value string
}

type CreatePolicyRuleRequestActionTypeEnum struct {
	PUT_RECORDS CreatePolicyRuleRequestActionType
	GET_RECORDS CreatePolicyRuleRequestActionType
}

func GetCreatePolicyRuleRequestActionTypeEnum() CreatePolicyRuleRequestActionTypeEnum {
	return CreatePolicyRuleRequestActionTypeEnum{
		PUT_RECORDS: CreatePolicyRuleRequestActionType{
			value: "putRecords",
		},
		GET_RECORDS: CreatePolicyRuleRequestActionType{
			value: "getRecords",
		},
	}
}

func (c CreatePolicyRuleRequestActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePolicyRuleRequestActionType) UnmarshalJSON(b []byte) error {
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

type CreatePolicyRuleRequestEffect struct {
	value string
}

type CreatePolicyRuleRequestEffectEnum struct {
	ACCEPT CreatePolicyRuleRequestEffect
}

func GetCreatePolicyRuleRequestEffectEnum() CreatePolicyRuleRequestEffectEnum {
	return CreatePolicyRuleRequestEffectEnum{
		ACCEPT: CreatePolicyRuleRequestEffect{
			value: "accept",
		},
	}
}

func (c CreatePolicyRuleRequestEffect) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePolicyRuleRequestEffect) UnmarshalJSON(b []byte) error {
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
