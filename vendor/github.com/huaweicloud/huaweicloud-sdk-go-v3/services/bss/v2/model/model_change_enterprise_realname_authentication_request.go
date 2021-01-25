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
type ChangeEnterpriseRealnameAuthenticationRequest struct {
	Body *ChangeEnterpriseRealnameAuthsReq `json:"body,omitempty"`
}

func (o ChangeEnterpriseRealnameAuthenticationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeEnterpriseRealnameAuthenticationRequest struct{}"
	}

	return strings.Join([]string{"ChangeEnterpriseRealnameAuthenticationRequest", string(data)}, " ")
}
