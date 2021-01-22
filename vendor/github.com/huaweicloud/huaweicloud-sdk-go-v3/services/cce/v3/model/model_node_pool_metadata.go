/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type NodePoolMetadata struct {
	// 节点名池名称。  > 命名规则： >  >  - 以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-50位，且不能以中划线(-)结尾。 >  >  - 不允许创建名为 DefaultPool 的节点池。
	Name string `json:"name"`
	// 节点池的uid。创建成功后自动生成，填写无效
	Uid *string `json:"uid,omitempty"`
}

func (o NodePoolMetadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NodePoolMetadata struct{}"
	}

	return strings.Join([]string{"NodePoolMetadata", string(data)}, " ")
}
