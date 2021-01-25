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
type CreateEnterpriseRealnameAuthenticationRequest struct {
	Body *ApplyEnterpriseRealnameAuthsReq `json:"body,omitempty"`
}

func (o CreateEnterpriseRealnameAuthenticationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEnterpriseRealnameAuthenticationRequest struct{}"
	}

	return strings.Join([]string{"CreateEnterpriseRealnameAuthenticationRequest", string(data)}, " ")
}
