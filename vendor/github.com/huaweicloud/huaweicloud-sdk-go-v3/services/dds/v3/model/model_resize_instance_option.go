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

type ResizeInstanceOption struct {
	// 对象类型。 - 对于集群实例，该参数为必选。变更mongos节点规格时，取值为“mongos”；变更shard组规格时，取值为“shard”。 - 对于副本集和单节点实例，不传该参数。
	TargetType *ResizeInstanceOptionTargetType `json:"target_type,omitempty"`
	// 待变更规格的节点ID或实例ID。 - 对于集群实例，变更mongos节点规格时，取值为mongos节点ID；变更shard组规格时，取值为shard组ID。 - 对于副本集实例，取值为相应的实例ID。 - 对于单节点实例，取值为相应的实例ID。
	TargetId string `json:"target_id"`
	// 变更至新规格的资源规格编码。
	TargetSpecCode string `json:"target_spec_code"`
}

func (o ResizeInstanceOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceOption struct{}"
	}

	return strings.Join([]string{"ResizeInstanceOption", string(data)}, " ")
}

type ResizeInstanceOptionTargetType struct {
	value string
}

type ResizeInstanceOptionTargetTypeEnum struct {
	MONGOS ResizeInstanceOptionTargetType
	SHARD  ResizeInstanceOptionTargetType
}

func GetResizeInstanceOptionTargetTypeEnum() ResizeInstanceOptionTargetTypeEnum {
	return ResizeInstanceOptionTargetTypeEnum{
		MONGOS: ResizeInstanceOptionTargetType{
			value: "mongos",
		},
		SHARD: ResizeInstanceOptionTargetType{
			value: "shard",
		},
	}
}

func (c ResizeInstanceOptionTargetType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResizeInstanceOptionTargetType) UnmarshalJSON(b []byte) error {
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
