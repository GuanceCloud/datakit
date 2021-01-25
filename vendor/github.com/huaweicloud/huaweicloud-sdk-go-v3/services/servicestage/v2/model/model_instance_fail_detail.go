/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 失败描述。  cluster_deleted,        // 集群被删除  cluster_unavailable,    // 集群不可用  cluster_inaccessible,   // 集群无法访问  namespace_deleted,      // 命名空间被删除  namespace_unavailable,  // 命名空间不可用  namespace_inaccessible, // 命名空间无法访问  resource_deleted,       // 资源已删除
type InstanceFailDetail struct {
	value string
}

type InstanceFailDetailEnum struct {
	CLUSTER_DELETED        InstanceFailDetail
	CLUSTER_UNAVAILABLE    InstanceFailDetail
	CLUSTER_INACCESSIBLE   InstanceFailDetail
	NAMESPACE_DELETED      InstanceFailDetail
	NAMESPACE_UNAVAILABLE  InstanceFailDetail
	NAMESPACE_INACCESSIBLE InstanceFailDetail
	RESOURCE_DELETED       InstanceFailDetail
}

func GetInstanceFailDetailEnum() InstanceFailDetailEnum {
	return InstanceFailDetailEnum{
		CLUSTER_DELETED: InstanceFailDetail{
			value: "cluster_deleted",
		},
		CLUSTER_UNAVAILABLE: InstanceFailDetail{
			value: "cluster_unavailable",
		},
		CLUSTER_INACCESSIBLE: InstanceFailDetail{
			value: "cluster_inaccessible",
		},
		NAMESPACE_DELETED: InstanceFailDetail{
			value: "namespace_deleted",
		},
		NAMESPACE_UNAVAILABLE: InstanceFailDetail{
			value: "namespace_unavailable",
		},
		NAMESPACE_INACCESSIBLE: InstanceFailDetail{
			value: "namespace_inaccessible",
		},
		RESOURCE_DELETED: InstanceFailDetail{
			value: "resource_deleted",
		},
	}
}

func (c InstanceFailDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceFailDetail) UnmarshalJSON(b []byte) error {
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
