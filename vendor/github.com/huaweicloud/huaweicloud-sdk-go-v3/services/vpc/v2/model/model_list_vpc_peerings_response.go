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
type ListVpcPeeringsResponse struct {
	// peering对象列表
	Peerings *[]VpcPeering `json:"peerings,omitempty"`
	// 分页信息
	PeeringsLinks  *[]NeutronPageLink `json:"peerings_links,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListVpcPeeringsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVpcPeeringsResponse struct{}"
	}

	return strings.Join([]string{"ListVpcPeeringsResponse", string(data)}, " ")
}
