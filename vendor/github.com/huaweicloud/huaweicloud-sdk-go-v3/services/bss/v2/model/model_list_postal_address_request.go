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
type ListPostalAddressRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

func (o ListPostalAddressRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPostalAddressRequest struct{}"
	}

	return strings.Join([]string{"ListPostalAddressRequest", string(data)}, " ")
}
