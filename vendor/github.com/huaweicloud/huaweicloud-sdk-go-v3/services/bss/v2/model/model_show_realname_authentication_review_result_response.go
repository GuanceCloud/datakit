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
type ShowRealnameAuthenticationReviewResultResponse struct {
	// |参数名称：实名认证审核结果，只有状态码为200并且已经提交过实名认证请求才返回：0：审核中1：不通过2：通过| |参数的约束及描述：实名认证审核结果，只有状态码为200并且已经提交过实名认证请求才返回：0：审核中1：不通过2：通过|
	ReviewResult *int32 `json:"review_result,omitempty"`
	// |参数名称：审批意见，只有状态码为200并且审核不通过才返回。| |参数约束及描述：审批意见，只有状态码为200并且审核不通过才返回。|
	Opinion        *string `json:"opinion,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowRealnameAuthenticationReviewResultResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRealnameAuthenticationReviewResultResponse struct{}"
	}

	return strings.Join([]string{"ShowRealnameAuthenticationReviewResultResponse", string(data)}, " ")
}
