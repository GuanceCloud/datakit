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
type ChangeEnterpriseRealnameAuthenticationResponse struct {
	// |参数名称：是否需要转人工审核，只有状态码为200才返回该参数：0：不需要1：需要| |参数的约束及描述：是否需要转人工审核，只有状态码为200才返回该参数：0：不需要1：需要|
	IsReview       *int32 `json:"is_review,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ChangeEnterpriseRealnameAuthenticationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeEnterpriseRealnameAuthenticationResponse struct{}"
	}

	return strings.Join([]string{"ChangeEnterpriseRealnameAuthenticationResponse", string(data)}, " ")
}
