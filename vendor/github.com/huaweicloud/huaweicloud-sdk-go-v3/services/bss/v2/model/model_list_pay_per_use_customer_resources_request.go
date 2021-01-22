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
type ListPayPerUseCustomerResourcesRequest struct {
	Body *QueryResourcesReq `json:"body,omitempty"`
}

func (o ListPayPerUseCustomerResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPayPerUseCustomerResourcesRequest struct{}"
	}

	return strings.Join([]string{"ListPayPerUseCustomerResourcesRequest", string(data)}, " ")
}
