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

// 转存目标的描述。
type ShowSinkTaskDetailRespObsDestinationDescriptor struct {
	// 消费启动策略：  - latest：从Topic最后端开始消费。  - earliest: 从Topic最前端消息开始消费。  默认是latest。
	ConsumerStrategy *string `json:"consumer_strategy,omitempty"`
	// 转储文件格式。目前只支持text格式。
	DestinationFileType *string `json:"destination_file_type,omitempty"`
	// 存储该通道数据的OBS桶名称。
	ObsBucketName *string `json:"obs_bucket_name,omitempty"`
	// 存储在obs的路径。
	ObsPath *string `json:"obs_path,omitempty"`
	// 将转储文件的生成时间使用“yyyy/MM/dd/HH/mm”格式生成分区字符串，用来定义写到OBS的Object文件所在的目录层次结构。   - N/A：置空，不使用日期时间目录。   - yyyy：年   - yyyy/MM：年/月   - yyyy/MM/dd：年/月/日   - yyyy/MM/dd/HH：年/月/日/时   - yyyy/MM/dd/HH/mm：年/月/日/时/分，例如：2017/11/10/14/49，目录结构就是“2017 > 11 > 10 > 14 > 49”，“2017”表示最外层文件夹。  默认值：空 > 数据转储成功后，存储的目录结构为“obs_bucket_path/file_prefix/partition_format”。默认时间是GMT+8 时间
	PartitionFormat *string `json:"partition_format,omitempty"`
	// 转储文件的记录分隔符，用于分隔写入转储文件的用户数据。 取值范围：   - 逗号“,”   - 分号“;”   - 竖线“|”   - 换行符“\\n”   - NULL  默认值：换行符“\\n”。
	RecordDelimiter *string `json:"record_delimiter,omitempty"`
	// 根据用户配置的时间，周期性的将数据导入OBS，若某个时间段内无数据，则此时间段不会生成打包文件。 取值范围：30～900 缺省值：300 单位：秒。 > 使用OBS通道转储流式数据时该参数为必选配置。
	DeliverTimeInterval *string `json:"deliver_time_interval,omitempty"`
	// 每个传输文件多大后就开始上传，单位为byte。 默认值5242880。
	ObsPartSize *string `json:"obs_part_size,omitempty"`
}

func (o ShowSinkTaskDetailRespObsDestinationDescriptor) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSinkTaskDetailRespObsDestinationDescriptor struct{}"
	}

	return strings.Join([]string{"ShowSinkTaskDetailRespObsDestinationDescriptor", string(data)}, " ")
}
