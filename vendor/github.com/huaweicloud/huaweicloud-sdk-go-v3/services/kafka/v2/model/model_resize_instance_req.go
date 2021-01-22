/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ResizeInstanceReq struct {
	// 规格变更后的规格ID。  若只扩展磁盘大小，则规格ID保持和原实例不变。  规格ID请参考[查询实例的扩容规格列表](https://support.huaweicloud.com/api-kafka/ShowInstanceExtendProductInfo.html)接口。
	NewSpecCode *string `json:"new_spec_code,omitempty"`
	// 规格变更后的消息存储空间，单位：GB。  若扩展实例基准带宽，则new_storage_space不能低于基准带宽规定的最小磁盘大小。  磁盘空间大小请参考[查询实例的扩容规格列表](https://support.huaweicloud.com/api-kafka/ShowInstanceExtendProductInfo.html)接口。
	NewStorageSpace *int32 `json:"new_storage_space,omitempty"`
}

func (o ResizeInstanceReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceReq struct{}"
	}

	return strings.Join([]string{"ResizeInstanceReq", string(data)}, " ")
}
