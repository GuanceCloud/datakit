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
type ListCustomerOnDemandResourcesRequest struct {
	XLanguage *string                            `json:"X-Language,omitempty"`
	Body      *QueryCustomerOnDemandResourcesReq `json:"body,omitempty"`
}

func (o ListCustomerOnDemandResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerOnDemandResourcesRequest struct{}"
	}

	return strings.Join([]string{"ListCustomerOnDemandResourcesRequest", string(data)}, " ")
}
