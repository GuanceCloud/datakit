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

// Response Object
type ListKmsTagsResponse struct {
	// 标签列表，key和value键值对的集合。  - key：表示标签键，一个密钥下最多包含10个key，key不能为空，不能重复，同一个key中value不能重复。key最大长度为36个字符。  - value：表示标签值。每个值最大长度43个字符，value之间为“与”的关系。
	Tags           *[]Tag `json:"tags,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListKmsTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKmsTagsResponse struct{}"
	}

	return strings.Join([]string{"ListKmsTagsResponse", string(data)}, " ")
}
