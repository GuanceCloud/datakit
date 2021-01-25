/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"
	"os"

	"strings"
)

type Record struct {
	// 用户上传数据时设置的partition_key。  说明：  上传数据时，如果传了partition_key参数，则下载数据时可返回此参数。如果上传数据时，未传partition_key参数，而是传入partition_id，则不返回partition_key。
	PartitionKey *string `json:"partition_key,omitempty"`
	// 该条数据的序列号。
	SequenceNumber *string `json:"sequence_number,omitempty"`
	// 下载的数据。  下载的数据为序列化之后的二进制数据（Base64编码后的字符串）。  比如下载数据接口返回的数据是“ZGF0YQ==”，“ZGF0YQ==”经过Base64解码之后是“data”。
	Data **os.File `json:"data,omitempty"`
	// 记录写入DIS的时间戳。
	Timestamp *int64 `json:"timestamp,omitempty"`
	// 时间戳类型。  - CreateTime：创建时间。
	TimestampType *string `json:"timestamp_type,omitempty"`
}

func (o Record) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Record struct{}"
	}

	return strings.Join([]string{"Record", string(data)}, " ")
}
