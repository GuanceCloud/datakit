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
type CreateSubCustomerResponse struct {
	// |参数名称：客户ID| |参数的约束及描述：只有成功或客户向伙伴授权发生异常（CBC.5025）时才会返回，且只允许最大长度64的字符串|
	DomainId *string `json:"domain_id,omitempty"`
	// |参数名称：用户登录名| |参数的约束及描述：只有成功的时候才会返回，且只允许最大长度64的字符串|
	DomainName     *string `json:"domain_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateSubCustomerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubCustomerResponse struct{}"
	}

	return strings.Join([]string{"CreateSubCustomerResponse", string(data)}, " ")
}
