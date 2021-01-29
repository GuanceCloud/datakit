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

// 重启实例时必填。
type InstanceActionRequestRestart struct {
	// 在线调试时必填。
	Restart *InstanceActionRequestRestartRestart `json:"restart,omitempty"`
}

func (o InstanceActionRequestRestart) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceActionRequestRestart struct{}"
	}

	return strings.Join([]string{"InstanceActionRequestRestart", string(data)}, " ")
}

type InstanceActionRequestRestartRestart struct {
	value string
}

type InstanceActionRequestRestartRestartEnum struct {
	TRUE InstanceActionRequestRestartRestart
}

func GetInstanceActionRequestRestartRestartEnum() InstanceActionRequestRestartRestartEnum {
	return InstanceActionRequestRestartRestartEnum{
		TRUE: InstanceActionRequestRestartRestart{
			value: "true",
		},
	}
}

func (c InstanceActionRequestRestartRestart) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceActionRequestRestartRestart) UnmarshalJSON(b []byte) error {
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
