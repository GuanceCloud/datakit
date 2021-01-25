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

type CceClusterNodeInformationMetadata struct {
	// 节点名称  > 修改节点名称后，弹性云服务器名称（虚拟机名称）会同步修改。 > > 命名规则：以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-56位，且不能以中划线(-)结尾。
	Name string `json:"name"`
}

func (o CceClusterNodeInformationMetadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CceClusterNodeInformationMetadata struct{}"
	}

	return strings.Join([]string{"CceClusterNodeInformationMetadata", string(data)}, " ")
}
