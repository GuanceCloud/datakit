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

type CreateStreamReq struct {
	// 通道名称。  通道名称由字母、数字、下划线和中划线组成，长度为1～64字符。
	StreamName string `json:"stream_name"`
	// 分区数量。  分区是DIS数据通道的基本吞吐量单位。
	PartitionCount int32 `json:"partition_count"`
	// 通道类型。  - COMMON：普通通道，表示1MB带宽。 - ADVANCED：高级通道，表示5MB带宽。
	StreamType *CreateStreamReqStreamType `json:"stream_type,omitempty"`
	// 源数据类型。  - BLOB：存储在数据库管理系统中的一组二进制数据。 - JSON：一种开放的文件格式，以易读的文字为基础，用来传输由属性值或者序列性的值组成的数据对象。 - CSV：纯文本形式存储的表格数据，分隔符默认采用逗号。  缺省值：BLOB。
	DataType *CreateStreamReqDataType `json:"data_type,omitempty"`
	// 数据保留时长。  取值范围：24~72。  单位：小时。  缺省值：24。  空表示使用缺省值。
	DataDuration *int32 `json:"data_duration,omitempty"`
	// 是否开启自动扩缩容。  - true：开启自动扩缩容。 - false：关闭自动扩缩容。  默认不开启。
	AutoScaleEnabled *bool `json:"auto_scale_enabled,omitempty"`
	// 当自动扩缩容启用时，自动缩容的最小分片数。
	AutoScaleMinPartitionCount *int64 `json:"auto_scale_min_partition_count,omitempty"`
	// 当自动扩缩容启用时，自动扩容的最大分片数。
	AutoScaleMaxPartitionCount *int32 `json:"auto_scale_max_partition_count,omitempty"`
	// 用于描述用户JSON、CSV格式的源数据结构，采用Avro Schema的语法描述。
	DataSchema    *string        `json:"data_schema,omitempty"`
	CsvProperties *CsvProperties `json:"csv_properties,omitempty"`
	// 数据的压缩类型，目前支持：  - snappy - gzip - zip  默认不压缩。
	CompressionFormat *CreateStreamReqCompressionFormat `json:"compression_format,omitempty"`
	// 通道标签列表。
	Tags *[]Tag `json:"tags,omitempty"`
	// 通道标签列表。
	SysTags *[]SysTag `json:"sys_tags,omitempty"`
}

func (o CreateStreamReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateStreamReq struct{}"
	}

	return strings.Join([]string{"CreateStreamReq", string(data)}, " ")
}

type CreateStreamReqStreamType struct {
	value string
}

type CreateStreamReqStreamTypeEnum struct {
	COMMON   CreateStreamReqStreamType
	ADVANCED CreateStreamReqStreamType
}

func GetCreateStreamReqStreamTypeEnum() CreateStreamReqStreamTypeEnum {
	return CreateStreamReqStreamTypeEnum{
		COMMON: CreateStreamReqStreamType{
			value: "COMMON",
		},
		ADVANCED: CreateStreamReqStreamType{
			value: "ADVANCED",
		},
	}
}

func (c CreateStreamReqStreamType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateStreamReqStreamType) UnmarshalJSON(b []byte) error {
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

type CreateStreamReqDataType struct {
	value string
}

type CreateStreamReqDataTypeEnum struct {
	BLOB CreateStreamReqDataType
	JSON CreateStreamReqDataType
	CSV  CreateStreamReqDataType
}

func GetCreateStreamReqDataTypeEnum() CreateStreamReqDataTypeEnum {
	return CreateStreamReqDataTypeEnum{
		BLOB: CreateStreamReqDataType{
			value: "BLOB",
		},
		JSON: CreateStreamReqDataType{
			value: "JSON",
		},
		CSV: CreateStreamReqDataType{
			value: "CSV",
		},
	}
}

func (c CreateStreamReqDataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateStreamReqDataType) UnmarshalJSON(b []byte) error {
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

type CreateStreamReqCompressionFormat struct {
	value string
}

type CreateStreamReqCompressionFormatEnum struct {
	SNAPPY CreateStreamReqCompressionFormat
	GZIP   CreateStreamReqCompressionFormat
	ZIP    CreateStreamReqCompressionFormat
}

func GetCreateStreamReqCompressionFormatEnum() CreateStreamReqCompressionFormatEnum {
	return CreateStreamReqCompressionFormatEnum{
		SNAPPY: CreateStreamReqCompressionFormat{
			value: "snappy",
		},
		GZIP: CreateStreamReqCompressionFormat{
			value: "gzip",
		},
		ZIP: CreateStreamReqCompressionFormat{
			value: "zip",
		},
	}
}

func (c CreateStreamReqCompressionFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateStreamReqCompressionFormat) UnmarshalJSON(b []byte) error {
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
