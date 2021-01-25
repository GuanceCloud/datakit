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

// Response Object
type DescribeStreamResponse struct {
	// 通道名称。
	StreamName *string `json:"stream_name,omitempty"`
	// 通道创建的时间，13位时间戳。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 通道最近一次修改的时间，13位时间戳。
	LastModifiedTime *int64 `json:"last_modified_time,omitempty"`
	// 通道的当前状态。  - CREATING：创建中 - RUNNING：运行中 - TERMINATING：删除中 - TERMINATED：已删除
	Status *DescribeStreamResponseStatus `json:"status,omitempty"`
	// 通道类型。  - COMMON：普通通道，表示1MB带宽。 - ADVANCED：高级通道，表示5MB带宽。
	StreamType *DescribeStreamResponseStreamType `json:"stream_type,omitempty"`
	// A list of partitions that comprise the DIS stream.
	Partitions *[]PartitionResult `json:"partitions,omitempty"`
	// 是否还有更多满足请求条件的分区。  - 是：true。 - 否：false。
	HasMorePartitions *bool `json:"has_more_partitions,omitempty"`
	// 数据保留时长，单位是小时。
	RetentionPeriod *int32 `json:"retention_period,omitempty"`
	// 通道唯一标示符。
	StreamId *string `json:"stream_id,omitempty"`
	// 源数据类型。  - BLOB：存储在数据库管理系统中的一组二进制数据。 - JSON：一种开放的文件格式，以易读的文字为基础，用来传输由属性值或者序列性的值组成的数据对象。 - CSV：纯文本形式存储的表格数据，分隔符默认采用逗号。  缺省值：BLOB。
	DataType *DescribeStreamResponseDataType `json:"data_type,omitempty"`
	// 用于描述用户JSON、CSV格式的源数据结构，采用Avro Schema的语法描述。Avro介绍您也可以点击[这里](http://avro.apache.org/docs/current/#schemas)查看。
	DataSchema *string `json:"data_schema,omitempty"`
	// 数据的压缩类型，目前支持：  - snappy - gzip - zip  默认不压缩。
	CompressionFormat *DescribeStreamResponseCompressionFormat `json:"compression_format,omitempty"`
	CsvProperties     *CsvProperties                           `json:"csv_properties,omitempty"`
	// 可写分区总数量（只包含ACTIVE状态的分区）。
	WritablePartitionCount *int32 `json:"writable_partition_count,omitempty"`
	// 可读分区总数量（包含ACTIVE与DELETED状态的分区）。
	ReadablePartitionCount *int32 `json:"readable_partition_count,omitempty"`
	// 扩缩容操作记录列表。
	UpdatePartitionCounts *[]UpdatePartitionCount `json:"update_partition_counts,omitempty"`
	// 通道的标签列表。
	Tags *[]Tag `json:"tags,omitempty"`
	// 是否开启自动扩缩容。  - true：开启自动扩缩容。 - false：关闭自动扩缩容。  默认不开启。
	AutoScaleEnabled *bool `json:"auto_scale_enabled,omitempty"`
	// 当自动扩缩容启用时，自动缩容的最小分片数。
	AutoScaleMinPartitionCount *int32 `json:"auto_scale_min_partition_count,omitempty"`
	// 当自动扩缩容启用时，自动扩容的最大分片数。
	AutoScaleMaxPartitionCount *int32 `json:"auto_scale_max_partition_count,omitempty"`
	HttpStatusCode             int    `json:"-"`
}

func (o DescribeStreamResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DescribeStreamResponse struct{}"
	}

	return strings.Join([]string{"DescribeStreamResponse", string(data)}, " ")
}

type DescribeStreamResponseStatus struct {
	value string
}

type DescribeStreamResponseStatusEnum struct {
	CREATING    DescribeStreamResponseStatus
	RUNNING     DescribeStreamResponseStatus
	TERMINATING DescribeStreamResponseStatus
	FROZEN      DescribeStreamResponseStatus
}

func GetDescribeStreamResponseStatusEnum() DescribeStreamResponseStatusEnum {
	return DescribeStreamResponseStatusEnum{
		CREATING: DescribeStreamResponseStatus{
			value: "CREATING",
		},
		RUNNING: DescribeStreamResponseStatus{
			value: "RUNNING",
		},
		TERMINATING: DescribeStreamResponseStatus{
			value: "TERMINATING",
		},
		FROZEN: DescribeStreamResponseStatus{
			value: "FROZEN",
		},
	}
}

func (c DescribeStreamResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DescribeStreamResponseStatus) UnmarshalJSON(b []byte) error {
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

type DescribeStreamResponseStreamType struct {
	value string
}

type DescribeStreamResponseStreamTypeEnum struct {
	COMMON   DescribeStreamResponseStreamType
	ADVANCED DescribeStreamResponseStreamType
}

func GetDescribeStreamResponseStreamTypeEnum() DescribeStreamResponseStreamTypeEnum {
	return DescribeStreamResponseStreamTypeEnum{
		COMMON: DescribeStreamResponseStreamType{
			value: "COMMON",
		},
		ADVANCED: DescribeStreamResponseStreamType{
			value: "ADVANCED",
		},
	}
}

func (c DescribeStreamResponseStreamType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DescribeStreamResponseStreamType) UnmarshalJSON(b []byte) error {
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

type DescribeStreamResponseDataType struct {
	value string
}

type DescribeStreamResponseDataTypeEnum struct {
	BLOB DescribeStreamResponseDataType
	JSON DescribeStreamResponseDataType
	CSV  DescribeStreamResponseDataType
}

func GetDescribeStreamResponseDataTypeEnum() DescribeStreamResponseDataTypeEnum {
	return DescribeStreamResponseDataTypeEnum{
		BLOB: DescribeStreamResponseDataType{
			value: "BLOB",
		},
		JSON: DescribeStreamResponseDataType{
			value: "JSON",
		},
		CSV: DescribeStreamResponseDataType{
			value: "CSV",
		},
	}
}

func (c DescribeStreamResponseDataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DescribeStreamResponseDataType) UnmarshalJSON(b []byte) error {
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

type DescribeStreamResponseCompressionFormat struct {
	value string
}

type DescribeStreamResponseCompressionFormatEnum struct {
	SNAPPY DescribeStreamResponseCompressionFormat
	GZIP   DescribeStreamResponseCompressionFormat
	ZIP    DescribeStreamResponseCompressionFormat
}

func GetDescribeStreamResponseCompressionFormatEnum() DescribeStreamResponseCompressionFormatEnum {
	return DescribeStreamResponseCompressionFormatEnum{
		SNAPPY: DescribeStreamResponseCompressionFormat{
			value: "snappy",
		},
		GZIP: DescribeStreamResponseCompressionFormat{
			value: "gzip",
		},
		ZIP: DescribeStreamResponseCompressionFormat{
			value: "zip",
		},
	}
}

func (c DescribeStreamResponseCompressionFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DescribeStreamResponseCompressionFormat) UnmarshalJSON(b []byte) error {
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
