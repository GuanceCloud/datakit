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

type PartitionResult struct {
	// 分区的当前状态。  - CREATING：创建中 - ACTIVE：可用 - DELETED：删除中 - EXPIRED：已过期
	Status *PartitionResultStatus `json:"status,omitempty"`
	// 分区的唯一标识符。
	PartitionId *string `json:"partition_id,omitempty"`
	// 分区的可能哈希键值范围。
	HashRange *string `json:"hash_range,omitempty"`
	// 分区的序列号范围。
	SequenceNumberRange *string `json:"sequence_number_range,omitempty"`
	// 父分区。
	ParentPartitions *string `json:"parent_partitions,omitempty"`
}

func (o PartitionResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PartitionResult struct{}"
	}

	return strings.Join([]string{"PartitionResult", string(data)}, " ")
}

type PartitionResultStatus struct {
	value string
}

type PartitionResultStatusEnum struct {
	CREATING PartitionResultStatus
	ACTIVE   PartitionResultStatus
	DELETED  PartitionResultStatus
	EXPIRED  PartitionResultStatus
}

func GetPartitionResultStatusEnum() PartitionResultStatusEnum {
	return PartitionResultStatusEnum{
		CREATING: PartitionResultStatus{
			value: "CREATING",
		},
		ACTIVE: PartitionResultStatus{
			value: "ACTIVE",
		},
		DELETED: PartitionResultStatus{
			value: "DELETED",
		},
		EXPIRED: PartitionResultStatus{
			value: "EXPIRED",
		},
	}
}

func (c PartitionResultStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PartitionResultStatus) UnmarshalJSON(b []byte) error {
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
