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

// Request Object
type SendSmsVerificationCodeRequest struct {
	Body *SendSmVerificationCodeReq `json:"body,omitempty"`
}

func (o SendSmsVerificationCodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendSmsVerificationCodeRequest struct{}"
	}

	return strings.Join([]string{"SendSmsVerificationCodeRequest", string(data)}, " ")
}
