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

type RestartInstanceRequestBody struct {
	// 待重启对象的类型。 - 重启集群实例下的节点时，该参数必选。取值为“mongos”、“shard”、或“config”。 - 重启整个实例时，不传该参数。
	TargetType *RestartInstanceRequestBodyTargetType `json:"target_type,omitempty"`
	// 待重启对象的ID。 - 重启集群实例下的节点时，对于mongos节点，取值为mongos节点ID，对于shard和config组，取值为shard和config组ID。 - 重启整个实例时，取值为实例ID。
	TargetId string `json:"target_id"`
}

func (o RestartInstanceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestartInstanceRequestBody struct{}"
	}

	return strings.Join([]string{"RestartInstanceRequestBody", string(data)}, " ")
}

type RestartInstanceRequestBodyTargetType struct {
	value string
}

type RestartInstanceRequestBodyTargetTypeEnum struct {
	MONGOS RestartInstanceRequestBodyTargetType
	SHARD  RestartInstanceRequestBodyTargetType
	CONFIG RestartInstanceRequestBodyTargetType
}

func GetRestartInstanceRequestBodyTargetTypeEnum() RestartInstanceRequestBodyTargetTypeEnum {
	return RestartInstanceRequestBodyTargetTypeEnum{
		MONGOS: RestartInstanceRequestBodyTargetType{
			value: "mongos",
		},
		SHARD: RestartInstanceRequestBodyTargetType{
			value: "shard",
		},
		CONFIG: RestartInstanceRequestBodyTargetType{
			value: "config",
		},
	}
}

func (c RestartInstanceRequestBodyTargetType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RestartInstanceRequestBodyTargetType) UnmarshalJSON(b []byte) error {
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
