/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ListKmsByTagsRequestBody struct {
	// 查询记录数（“action”为“count”时，无需设置此参数），如果“action”为“filter”，默认为“10”。 limit的取值范围为“1-1000”。
	Limit *string `json:"limit,omitempty"`
	// 索引位置。从offset指定的下一条数据开始查询。查询第一页数据时，将查询前一页数据时响应体中的值带入此参数（“action”为“count”时，无需设置此参数）。如果“action”为“filter”，offset默认为“0”。 offset必须为数字，不能为负数。
	Offset *string `json:"offset,omitempty"`
	// 操作标识（可设置为“filter”或者“count”）。  - filter：表示过滤。  - count：表示查询总条数。
	Action *string `json:"action,omitempty"`
	// 标签列表，key和value键值对的集合。  - key：表示标签键，一个密钥下最多包含10个key，key不能为空，不能重复，同一个key中value不能重复。key最大长度为36个字符。  - value：表示标签值。每个值最大长度43个字符，value之间为“与”的关系。
	Tags *[]Tag `json:"tags,omitempty"`
	// 搜索字段。  - key为要匹配的字段，例如：resource_name等。  - value为匹配的值，最大长度为255个字符，不能为空。
	Matches *[]TagItem `json:"matches,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o ListKmsByTagsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKmsByTagsRequestBody struct{}"
	}

	return strings.Join([]string{"ListKmsByTagsRequestBody", string(data)}, " ")
}
