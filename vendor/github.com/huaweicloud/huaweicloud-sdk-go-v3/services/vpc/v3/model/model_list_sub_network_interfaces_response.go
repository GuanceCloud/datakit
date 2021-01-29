/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListSubNetworkInterfacesResponse struct {
	// 1、功能说明：请求ID 2、取值范围：标准UUID 3、约束：N/A 4、默认值：N/A 5、权限：N/A
	RequestId *string `json:"request_id,omitempty"`
	// 1、功能说明：辅助弹性网卡查询对象 2、取值范围：N/A 3、约束：N/A 4、默认值：N/A 5、权限：N/A
	SubNetworkInterfaces *[]SubNetworkInterface `json:"sub_network_interfaces,omitempty"`
	PageInfo             *PageInfo              `json:"page_info,omitempty"`
	HttpStatusCode       int                    `json:"-"`
}

func (o ListSubNetworkInterfacesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubNetworkInterfacesResponse struct{}"
	}

	return strings.Join([]string{"ListSubNetworkInterfacesResponse", string(data)}, " ")
}
