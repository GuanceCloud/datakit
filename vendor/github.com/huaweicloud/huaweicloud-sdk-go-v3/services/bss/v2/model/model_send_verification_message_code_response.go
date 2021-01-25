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
type SendVerificationMessageCodeResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o SendVerificationMessageCodeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendVerificationMessageCodeResponse struct{}"
	}

	return strings.Join([]string{"SendVerificationMessageCodeResponse", string(data)}, " ")
}
