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
type SendSmsVerificationCodeResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o SendSmsVerificationCodeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendSmsVerificationCodeResponse struct{}"
	}

	return strings.Join([]string{"SendSmsVerificationCodeResponse", string(data)}, " ")
}
