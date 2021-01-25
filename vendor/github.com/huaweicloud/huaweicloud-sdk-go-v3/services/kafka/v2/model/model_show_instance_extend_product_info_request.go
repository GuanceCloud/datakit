/*
 * Kafka
 *
 * Kafka Document API
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
type ShowInstanceExtendProductInfoRequest struct {
	ProjectId  string                                     `json:"project_id"`
	InstanceId string                                     `json:"instance_id"`
	Type       ShowInstanceExtendProductInfoRequestType   `json:"type"`
	Engine     ShowInstanceExtendProductInfoRequestEngine `json:"engine"`
}

func (o ShowInstanceExtendProductInfoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceExtendProductInfoRequest struct{}"
	}

	return strings.Join([]string{"ShowInstanceExtendProductInfoRequest", string(data)}, " ")
}

type ShowInstanceExtendProductInfoRequestType struct {
	value string
}

type ShowInstanceExtendProductInfoRequestTypeEnum struct {
	ADVANCED ShowInstanceExtendProductInfoRequestType
	PLATINUM ShowInstanceExtendProductInfoRequestType
	DEC      ShowInstanceExtendProductInfoRequestType
	EXP      ShowInstanceExtendProductInfoRequestType
}

func GetShowInstanceExtendProductInfoRequestTypeEnum() ShowInstanceExtendProductInfoRequestTypeEnum {
	return ShowInstanceExtendProductInfoRequestTypeEnum{
		ADVANCED: ShowInstanceExtendProductInfoRequestType{
			value: "advanced",
		},
		PLATINUM: ShowInstanceExtendProductInfoRequestType{
			value: "platinum",
		},
		DEC: ShowInstanceExtendProductInfoRequestType{
			value: "dec",
		},
		EXP: ShowInstanceExtendProductInfoRequestType{
			value: "exp",
		},
	}
}

func (c ShowInstanceExtendProductInfoRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowInstanceExtendProductInfoRequestType) UnmarshalJSON(b []byte) error {
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

type ShowInstanceExtendProductInfoRequestEngine struct {
	value string
}

type ShowInstanceExtendProductInfoRequestEngineEnum struct {
	KAFKA ShowInstanceExtendProductInfoRequestEngine
}

func GetShowInstanceExtendProductInfoRequestEngineEnum() ShowInstanceExtendProductInfoRequestEngineEnum {
	return ShowInstanceExtendProductInfoRequestEngineEnum{
		KAFKA: ShowInstanceExtendProductInfoRequestEngine{
			value: "kafka",
		},
	}
}

func (c ShowInstanceExtendProductInfoRequestEngine) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowInstanceExtendProductInfoRequestEngine) UnmarshalJSON(b []byte) error {
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
