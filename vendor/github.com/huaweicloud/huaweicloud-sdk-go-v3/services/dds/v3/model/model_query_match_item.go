/*
 * DDS
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

type QueryMatchItem struct {
	// 取值为“instance_name”或“instance_id”，分别表示按实例名称或按实例ID匹配查询。
	Key QueryMatchItemKey `json:"key"`
	// 待匹配的实例名称或实例ID。
	Value string `json:"value"`
}

func (o QueryMatchItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryMatchItem struct{}"
	}

	return strings.Join([]string{"QueryMatchItem", string(data)}, " ")
}

type QueryMatchItemKey struct {
	value string
}

type QueryMatchItemKeyEnum struct {
	INSTANCE_NAME QueryMatchItemKey
	INSTANCE_ID   QueryMatchItemKey
}

func GetQueryMatchItemKeyEnum() QueryMatchItemKeyEnum {
	return QueryMatchItemKeyEnum{
		INSTANCE_NAME: QueryMatchItemKey{
			value: "instance_name",
		},
		INSTANCE_ID: QueryMatchItemKey{
			value: "instance_id",
		},
	}
}

func (c QueryMatchItemKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *QueryMatchItemKey) UnmarshalJSON(b []byte) error {
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
