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
type ListVpcRoutesResponse struct {
	// route对象列表
	Routes *[]VpcRoute `json:"routes,omitempty"`
	// 分页信息
	RoutesLinks    *[]NeutronPageLink `json:"routes_links,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListVpcRoutesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVpcRoutesResponse struct{}"
	}

	return strings.Join([]string{"ListVpcRoutesResponse", string(data)}, " ")
}
