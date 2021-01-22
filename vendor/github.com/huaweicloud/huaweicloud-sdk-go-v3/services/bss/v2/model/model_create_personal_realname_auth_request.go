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
type CreatePersonalRealnameAuthRequest struct {
	Body *ApplyIndividualRealnameAuthsReq `json:"body,omitempty"`
}

func (o CreatePersonalRealnameAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePersonalRealnameAuthRequest struct{}"
	}

	return strings.Join([]string{"CreatePersonalRealnameAuthRequest", string(data)}, " ")
}
