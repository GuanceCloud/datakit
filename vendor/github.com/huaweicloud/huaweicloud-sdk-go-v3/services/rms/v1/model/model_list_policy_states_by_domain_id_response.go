/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListPolicyStatesByDomainIdResponse struct {
	// 合规结果查询返回值
	Value          *[]PolicyState `json:"value,omitempty"`
	PageInfo       *PageInfo      `json:"page_info,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListPolicyStatesByDomainIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPolicyStatesByDomainIdResponse struct{}"
	}

	return strings.Join([]string{"ListPolicyStatesByDomainIdResponse", string(data)}, " ")
}
