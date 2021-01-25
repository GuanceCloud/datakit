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

// 实例规格详情。
type CreateInstanceFlavorOption struct {
	// 节点类型。 取值：   - 集群实例包含mongos、shard和config节点，各节点下该参数取值分别为“mongos”、“shard”和“config”。   - 副本集实例下该参数取值为“replica”。   - 单节点实例下该参数取值为“single”。
	Type CreateInstanceFlavorOptionType `json:"type"`
	// 节点数量。 取值：   - 集群实例下“mongos”类型的节点数量可取2~16。   - 集群实例下“shard”类型的组数量可取2~16。   - “shard”类型的组数量可取2~16，恢复到新实例不传该参数。   - “config”类型的组数量只能取1。   - “replica”类型的组数量只能取1。   - “single”类型的节点数量只能取1。
	Num int32 `json:"num"`
	// 磁盘类型。 取值：ULTRAHIGH，表示SSD。   - 对于集群实例的shard和config节点、副本集、以及单节点实例，该参数有效。mongos节点不涉及选择磁盘，该参数无意义。   - 恢复到新实例，不传该参数。
	Storage *string `json:"storage,omitempty"`
	// 磁盘大小。 取值：必须为10的整数倍。单位为GB。   - 对于集群实例，shard组可取10GB~2000GB，config组仅可取20GB。mongos节点不涉及选择磁盘，该参数无意义。   - 对于副本集实例，可取10GB~2000GB。   - 对于单节点实例，可取10GB~1000GB。
	Size *int32 `json:"size,omitempty"`
	// 资源规格编码
	SpecCode string `json:"spec_code"`
}

func (o CreateInstanceFlavorOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceFlavorOption struct{}"
	}

	return strings.Join([]string{"CreateInstanceFlavorOption", string(data)}, " ")
}

type CreateInstanceFlavorOptionType struct {
	value string
}

type CreateInstanceFlavorOptionTypeEnum struct {
	MONGOS  CreateInstanceFlavorOptionType
	SHARD   CreateInstanceFlavorOptionType
	CONFIG  CreateInstanceFlavorOptionType
	REPLICA CreateInstanceFlavorOptionType
	SINGLE  CreateInstanceFlavorOptionType
}

func GetCreateInstanceFlavorOptionTypeEnum() CreateInstanceFlavorOptionTypeEnum {
	return CreateInstanceFlavorOptionTypeEnum{
		MONGOS: CreateInstanceFlavorOptionType{
			value: "mongos",
		},
		SHARD: CreateInstanceFlavorOptionType{
			value: "shard",
		},
		CONFIG: CreateInstanceFlavorOptionType{
			value: "config",
		},
		REPLICA: CreateInstanceFlavorOptionType{
			value: "replica",
		},
		SINGLE: CreateInstanceFlavorOptionType{
			value: "single",
		},
	}
}

func (c CreateInstanceFlavorOptionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateInstanceFlavorOptionType) UnmarshalJSON(b []byte) error {
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
