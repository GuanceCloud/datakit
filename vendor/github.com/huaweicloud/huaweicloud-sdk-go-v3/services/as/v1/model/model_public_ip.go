/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 配置云服务器的弹性IP信息
type PublicIp struct {
	Eip *Eip `json:"eip"`
}

func (o PublicIp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublicIp struct{}"
	}

	return strings.Join([]string{"PublicIp", string(data)}, " ")
}
