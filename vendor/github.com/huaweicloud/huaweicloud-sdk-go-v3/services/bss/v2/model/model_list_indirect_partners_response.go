/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListIndirectPartnersResponse struct {
	// |参数名称：符合条件的记录个数，只有成功的时候出现| |参数的约束及描述：符合条件的记录个数，只有成功的时候出现|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：二级渠道邀请记录列表| |参数约束以及描述：二级渠道邀请记录列表|
	IndirectPartners *[]IndirectPartnerInfo `json:"indirect_partners,omitempty"`
	HttpStatusCode   int                    `json:"-"`
}

func (o ListIndirectPartnersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIndirectPartnersResponse struct{}"
	}

	return strings.Join([]string{"ListIndirectPartnersResponse", string(data)}, " ")
}
