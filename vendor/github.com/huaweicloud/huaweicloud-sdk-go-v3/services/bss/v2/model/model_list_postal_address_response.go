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
type ListPostalAddressResponse struct {
	// |参数名称：查询个数，成功的时候返回| |参数的约束及描述：查询个数，成功的时候返回|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：邮寄地址| |参数约束以及描述：邮寄地址|
	PostalAddress  *[]CustomerPostalAddressV2 `json:"postal_address,omitempty"`
	HttpStatusCode int                        `json:"-"`
}

func (o ListPostalAddressResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPostalAddressResponse struct{}"
	}

	return strings.Join([]string{"ListPostalAddressResponse", string(data)}, " ")
}
