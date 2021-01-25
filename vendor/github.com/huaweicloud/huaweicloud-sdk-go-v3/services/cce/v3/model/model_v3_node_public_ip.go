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

type V3NodePublicIp struct {
	// 要动态创建的弹性IP个数。 > count参数与eip参数必须同时配置。
	Count *int32         `json:"count,omitempty"`
	Eip   *V3NodeEipSpec `json:"eip,omitempty"`
	// 已有的弹性IP的ID列表。数量不得大于待创建节点数 > 若已配置ids参数，则无需配置count和eip参数
	Ids *[]string `json:"ids,omitempty"`
}

func (o V3NodePublicIp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V3NodePublicIp struct{}"
	}

	return strings.Join([]string{"V3NodePublicIp", string(data)}, " ")
}
