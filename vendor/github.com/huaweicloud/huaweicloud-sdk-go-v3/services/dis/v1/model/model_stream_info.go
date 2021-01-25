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

type StreamInfo struct {
	// 通道名称。
	StreamName *string `json:"stream_name,omitempty"`
	// 通道创建的时间，13位时间戳。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 数据保留时长，单位是小时。
	RetentionPeriod *int32 `json:"retention_period,omitempty"`
	// 通道的当前状态。  - CREATING：创建中 - RUNNING：运行中 - TERMINATING：删除中 - TERMINATED：已删除
	Status *StreamInfoStatus `json:"status,omitempty"`
	// 通道类型。  - COMMON：普通通道，表示1MB带宽。 - ADVANCED：高级通道，表示5MB带宽。
	StreamType *StreamInfoStreamType `json:"stream_type,omitempty"`
	// 源数据类型。  - BLOB：存储在数据库管理系统中的一组二进制数据。 - JSON：一种开放的文件格式，以易读的文字为基础，用来传输由属性值或者序列性的值组成的数据对象。 - CSV：纯文本形式存储的表格数据，分隔符默认采用逗号。  缺省值：BLOB。
	DataType *StreamInfoDataType `json:"data_type,omitempty"`
	// 分区数量。  分区是DIS数据通道的基本吞吐量单位。
	PartitionCount *int32 `json:"partition_count,omitempty"`
	// List of tags for the newly created DIS stream.
	Tags *[]Tag `json:"tags,omitempty"`
	// 是否开启自动扩缩容。  - true：开启自动扩缩容。 - false：关闭自动扩缩容。  默认不开启。
	AutoScaleEnabled *bool `json:"auto_scale_enabled,omitempty"`
	// 当自动扩缩容启用时，自动缩容的最小分片数。
	AutoScaleMinPartitionCount *int32 `json:"auto_scale_min_partition_count,omitempty"`
	// 当自动扩缩容启用时，自动扩容的最大分片数。
	AutoScaleMaxPartitionCount *int32 `json:"auto_scale_max_partition_count,omitempty"`
}

func (o StreamInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StreamInfo struct{}"
	}

	return strings.Join([]string{"StreamInfo", string(data)}, " ")
}

type StreamInfoStatus struct {
	value string
}

type StreamInfoStatusEnum struct {
	CREATING    StreamInfoStatus
	RUNNING     StreamInfoStatus
	TERMINATING StreamInfoStatus
	FROZEN      StreamInfoStatus
}

func GetStreamInfoStatusEnum() StreamInfoStatusEnum {
	return StreamInfoStatusEnum{
		CREATING: StreamInfoStatus{
			value: "CREATING",
		},
		RUNNING: StreamInfoStatus{
			value: "RUNNING",
		},
		TERMINATING: StreamInfoStatus{
			value: "TERMINATING",
		},
		FROZEN: StreamInfoStatus{
			value: "FROZEN",
		},
	}
}

func (c StreamInfoStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StreamInfoStatus) UnmarshalJSON(b []byte) error {
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

type StreamInfoStreamType struct {
	value string
}

type StreamInfoStreamTypeEnum struct {
	COMMON   StreamInfoStreamType
	ADVANCED StreamInfoStreamType
}

func GetStreamInfoStreamTypeEnum() StreamInfoStreamTypeEnum {
	return StreamInfoStreamTypeEnum{
		COMMON: StreamInfoStreamType{
			value: "COMMON",
		},
		ADVANCED: StreamInfoStreamType{
			value: "ADVANCED",
		},
	}
}

func (c StreamInfoStreamType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StreamInfoStreamType) UnmarshalJSON(b []byte) error {
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

type StreamInfoDataType struct {
	value string
}

type StreamInfoDataTypeEnum struct {
	BLOB StreamInfoDataType
	JSON StreamInfoDataType
	CSV  StreamInfoDataType
}

func GetStreamInfoDataTypeEnum() StreamInfoDataTypeEnum {
	return StreamInfoDataTypeEnum{
		BLOB: StreamInfoDataType{
			value: "BLOB",
		},
		JSON: StreamInfoDataType{
			value: "JSON",
		},
		CSV: StreamInfoDataType{
			value: "CSV",
		},
	}
}

func (c StreamInfoDataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StreamInfoDataType) UnmarshalJSON(b []byte) error {
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
