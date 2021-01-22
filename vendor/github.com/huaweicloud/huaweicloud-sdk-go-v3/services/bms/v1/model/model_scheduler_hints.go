/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// schedulerHints字段数据结构说明
type SchedulerHints struct {
	// 是否在专属云中创建裸金属服务器，参数值为share或dedicate。约束：该值不传时默认为share。在专属云中创建裸金属服务器时，必须指定该字段为dedicate。
	DecBaremetal *[]SchedulerHintsDecBaremetal `json:"dec_baremetal,omitempty"`
}

func (o SchedulerHints) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SchedulerHints struct{}"
	}

	return strings.Join([]string{"SchedulerHints", string(data)}, " ")
}

type SchedulerHintsDecBaremetal struct {
	value string
}

type SchedulerHintsDecBaremetalEnum struct {
	SHARE    SchedulerHintsDecBaremetal
	DEDICATE SchedulerHintsDecBaremetal
}

func GetSchedulerHintsDecBaremetalEnum() SchedulerHintsDecBaremetalEnum {
	return SchedulerHintsDecBaremetalEnum{
		SHARE: SchedulerHintsDecBaremetal{
			value: "share",
		},
		DEDICATE: SchedulerHintsDecBaremetal{
			value: "dedicate",
		},
	}
}

func (c SchedulerHintsDecBaremetal) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SchedulerHintsDecBaremetal) UnmarshalJSON(b []byte) error {
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
