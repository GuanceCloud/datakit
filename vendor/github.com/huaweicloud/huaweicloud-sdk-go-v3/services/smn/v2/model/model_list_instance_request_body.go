/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ListInstanceRequestBody struct {
	// 最多包含10个key，每个key最多包含10个value，结构体不能缺失。key不能为空或者空字符串。key不能重复，同一个key中value不能重复，不同key对应的资源之间为与的关系。
	Tags *[]ResourceTags `json:"tags,omitempty"`
	// 最多包含10个key，每个key最多包含10个value，结构体不能缺失。key不能为空或者空字符串。key不能重复，同一个key中value不能重复，不同key对应的资源之间为或的关系。
	TagsAny *[]ResourceTags `json:"tags_any,omitempty"`
	// 最多包含10个key，每个key最多包含10个value，结构体不能缺失。key不能为空或者空字符串。key不能重复，同一个key中value不能重复，不同key对应的资源之间为与非的关系。
	NotTags *[]ResourceTags `json:"not_tags,omitempty"`
	// 最多包含10个key，每个key最多包含10个value，结构体不能缺失。key不能为空或者空字符串。key不能重复，同一个key中value不能重复，不同key对应的资源之间为或非的关系。
	NotTagsAny *[]ResourceTags `json:"not_tags_any,omitempty"`
	// 索引位置， 从offset指定的下一条数据开始查询。 查询第一页数据时，不需要传入此参数，查询后续页码数据时，将查询前一页数据时响应体中的值带入此参数。  action为count时无此参数。  action为filter时，默认为0，必须为数字，且不能为负数。
	Offset *string `json:"offset,omitempty"`
	// 查询记录数。  action为count时无此参数。  action为filter时，默认为1000。limit最多为1000，不能为负数，最小值为1。
	Limit *string `json:"limit,omitempty"`
	// 操作标识（仅限于filter，count）：filter（过滤），count(查询总条数)。 为filter时表示分页查询，为count只需按照条件将总条数返回即可。
	Action string `json:"action"`
	// 搜索字段。  key为要匹配的字段，当前只支持resource_name。  value为匹配的值，当前为精确匹配。
	Matches *[]TagMatch `json:"matches,omitempty"`
}

func (o ListInstanceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstanceRequestBody struct{}"
	}

	return strings.Join([]string{"ListInstanceRequestBody", string(data)}, " ")
}
