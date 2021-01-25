/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DisassociateKeypairRequest struct {
	Body *DisassociateKeypairRequestBody `json:"body,omitempty"`
}

func (o DisassociateKeypairRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociateKeypairRequest struct{}"
	}

	return strings.Join([]string{"DisassociateKeypairRequest", string(data)}, " ")
}
