/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type PrePaidServerEip struct {
	// 弹性IP地址类型。
	Iptype      string                       `json:"iptype"`
	Bandwidth   *PrePaidServerEipBandwidth   `json:"bandwidth"`
	Extendparam *PrePaidServerEipExtendParam `json:"extendparam,omitempty"`
}

func (o PrePaidServerEip) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PrePaidServerEip struct{}"
	}

	return strings.Join([]string{"PrePaidServerEip", string(data)}, " ")
}
