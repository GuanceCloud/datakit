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
type ListEnterpriseSubCustomersRequest struct {
	SubCustomerAccountName *string `json:"sub_customer_account_name,omitempty"`
	SubCustomerDisplayName *string `json:"sub_customer_display_name,omitempty"`
	FuzzyQuery             *int32  `json:"fuzzy_query,omitempty"`
	Offset                 *int32  `json:"offset,omitempty"`
	Limit                  *int32  `json:"limit,omitempty"`
	OrgId                  *string `json:"org_id,omitempty"`
}

func (o ListEnterpriseSubCustomersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnterpriseSubCustomersRequest struct{}"
	}

	return strings.Join([]string{"ListEnterpriseSubCustomersRequest", string(data)}, " ")
}
