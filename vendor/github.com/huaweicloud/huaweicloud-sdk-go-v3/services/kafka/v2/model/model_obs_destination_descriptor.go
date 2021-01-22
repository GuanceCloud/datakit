/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 转存目标的描述。
type ObsDestinationDescriptor struct {
	// 转存的topic列表名称，支持多个topic同时放置，以逗号“,”分隔。同时支持正则表达式。 例如topic1,topic2。
	Topics string `json:"topics"`
	// 转存topic的正则表达式，与topics必须二选一，不能同时都设置或者“.*”。
	TopicsRegex *string `json:"topics_regex,omitempty"`
	// 转储启动偏移量：   - latest: 从Topic最后端开始消费。   - earliest: 从Topic最前端消息开始消费。  默认是latest。
	ConsumerStrategy ObsDestinationDescriptorConsumerStrategy `json:"consumer_strategy"`
	// 转储文件格式。当前只支持text。
	DestinationFileType ObsDestinationDescriptorDestinationFileType `json:"destination_file_type"`
	// 访问密钥AK。
	AccessKey string `json:"access_key"`
	// 访问密钥SK。
	SecretKey string `json:"secret_key"`
	// 存储该通道数据的OBS桶名称。
	ObsBucketName string `json:"obs_bucket_name"`
	// 存储在obs的路径，默认可以不填。 取值范围：英文字母、数字、下划线和斜杠，最大长度为50个字符。 默认配置为空。
	ObsPath *string `json:"obs_path,omitempty"`
	// 将转储文件的生成时间使用“yyyy/MM/dd/HH/mm”格式生成分区字符串，用来定义写到OBS的Object文件所在的目录层次结构。   - N/A：置空，不使用日期时间目录。   - yyyy：年   - yyyy/MM：年/月   - yyyy/MM/dd：年/月/日   - yyyy/MM/dd/HH：年/月/日/时   - yyyy/MM/dd/HH/mm：年/月/日/时/分，例如：2017/11/10/14/49，目录结构就是“2017 > 11 > 10 > 14 > 49”，“2017”表示最外层文件夹。  默认值：空 > 数据转储成功后，存储的目录结构为“obs_bucket_path/file_prefix/partition_format”。默认时间是GMT+8 时间
	PartitionFormat *string `json:"partition_format,omitempty"`
	// 转储文件的记录分隔符，用于分隔写入转储文件的用户数据。 取值范围：   - 逗号“,”   - 分号“;”   - 竖线“|”   - 换行符“\\n”   - NULL  默认值：换行符“\\n”。
	RecordDelimiter *string `json:"record_delimiter,omitempty"`
	// 根据用户配置的时间，周期性的将数据导入OBS，若某个时间段内无数据，则此时间段不会生成打包文件。 取值范围：30～900 单位：秒。 > 使用OBS通道转储流式数据时该参数为必选配置。
	DeliverTimeInterval string `json:"deliver_time_interval"`
}

func (o ObsDestinationDescriptor) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ObsDestinationDescriptor struct{}"
	}

	return strings.Join([]string{"ObsDestinationDescriptor", string(data)}, " ")
}

type ObsDestinationDescriptorConsumerStrategy struct {
	value string
}

type ObsDestinationDescriptorConsumerStrategyEnum struct {
	LATEST   ObsDestinationDescriptorConsumerStrategy
	EARLIEST ObsDestinationDescriptorConsumerStrategy
}

func GetObsDestinationDescriptorConsumerStrategyEnum() ObsDestinationDescriptorConsumerStrategyEnum {
	return ObsDestinationDescriptorConsumerStrategyEnum{
		LATEST: ObsDestinationDescriptorConsumerStrategy{
			value: "latest",
		},
		EARLIEST: ObsDestinationDescriptorConsumerStrategy{
			value: "earliest",
		},
	}
}

func (c ObsDestinationDescriptorConsumerStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ObsDestinationDescriptorConsumerStrategy) UnmarshalJSON(b []byte) error {
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

type ObsDestinationDescriptorDestinationFileType struct {
	value string
}

type ObsDestinationDescriptorDestinationFileTypeEnum struct {
	TEXT ObsDestinationDescriptorDestinationFileType
}

func GetObsDestinationDescriptorDestinationFileTypeEnum() ObsDestinationDescriptorDestinationFileTypeEnum {
	return ObsDestinationDescriptorDestinationFileTypeEnum{
		TEXT: ObsDestinationDescriptorDestinationFileType{
			value: "TEXT",
		},
	}
}

func (c ObsDestinationDescriptorDestinationFileType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ObsDestinationDescriptorDestinationFileType) UnmarshalJSON(b []byte) error {
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
