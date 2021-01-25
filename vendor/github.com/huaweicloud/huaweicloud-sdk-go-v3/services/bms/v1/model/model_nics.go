/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// nics字段数据结构说明
type Nics struct {
	// 裸金属服务器网卡所在的子网信息。需要指定vpcid对应VPC下已创建的子网（subnet）的网络ID（network_id），UUID格式。子网（subnet）的网络ID（network_id）可以从虚拟私有云控制台或者参考《虚拟私有云API参考》的“查询子网列表”章节获取。
	SubnetId string `json:"subnet_id"`
	// 创建裸金属服务器网卡的IP地址，IPv4格式,约束：不填或空字符串，默认在网络（network）对应的子网中自动分配一个未使用的IP作网卡的IP地址。若指定IP地址，该IP地址必须在网络（network）对应的子网的网段内，且未被使用。
	IpAddress *string `json:"ip_address,omitempty"`
}

func (o Nics) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Nics struct{}"
	}

	return strings.Join([]string{"Nics", string(data)}, " ")
}
