/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 函数vpc配置。
type FuncVpc struct {
	// 虚拟私有云名称。
	VpcName string `json:"vpc_name"`
	// 虚拟私有云唯一标识。
	VpcId string `json:"vpc_id"`
	// 子网名称。
	SubnetName string `json:"subnet_name"`
	// 子网编号。
	SubnetId string `json:"subnet_id"`
	// 子网掩码。
	Cidr string `json:"cidr"`
	// 网关。
	Gateway string `json:"gateway"`
}

func (o FuncVpc) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FuncVpc struct{}"
	}

	return strings.Join([]string{"FuncVpc", string(data)}, " ")
}
