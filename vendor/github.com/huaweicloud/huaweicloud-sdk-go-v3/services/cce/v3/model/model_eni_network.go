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

// ENI网络配置，创建集群指定使用Yangtse网络模式时必填。
type EniNetwork struct {
	// ENI子网CIDR
	EniSubnetCIDR string `json:"eniSubnetCIDR"`
	// eni子网ID
	EniSubnetId string `json:"eniSubnetId"`
}

func (o EniNetwork) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EniNetwork struct{}"
	}

	return strings.Join([]string{"EniNetwork", string(data)}, " ")
}
