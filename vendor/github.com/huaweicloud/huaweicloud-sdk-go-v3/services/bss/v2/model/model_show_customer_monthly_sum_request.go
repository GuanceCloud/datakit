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
type ShowCustomerMonthlySumRequest struct {
	BillCycle           string  `json:"bill_cycle"`
	ServiceTypeCode     *string `json:"service_type_code,omitempty"`
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	Offset              *int32  `json:"offset,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
	Method              *string `json:"method,omitempty"`
	SubCustomerId       *string `json:"sub_customer_id,omitempty"`
}

func (o ShowCustomerMonthlySumRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerMonthlySumRequest struct{}"
	}

	return strings.Join([]string{"ShowCustomerMonthlySumRequest", string(data)}, " ")
}
