/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 标签
type TagReq struct {
	// 键。最大长度127个unicode字符。 key不能为空。(搜索时不对此参数做校验)
	Key string `json:"key"`
	// 值列表。每个值最大长度255个unicode字符，如果values为空列表，则表示any_value。value之间为或的关系。
	Values []string `json:"values"`
}

func (o TagReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagReq struct{}"
	}

	return strings.Join([]string{"TagReq", string(data)}, " ")
}
