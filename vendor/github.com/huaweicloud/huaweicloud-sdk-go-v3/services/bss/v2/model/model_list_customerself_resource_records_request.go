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
type ListCustomerselfResourceRecordsRequest struct {
	XLanguage           *string `json:"X-Language,omitempty"`
	Cycle               string  `json:"cycle"`
	CloudServiceType    *string `json:"cloud_service_type,omitempty"`
	Region              *string `json:"region,omitempty"`
	ChargeMode          *string `json:"charge_mode,omitempty"`
	BillType            *int32  `json:"bill_type,omitempty"`
	Offset              *int32  `json:"offset,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
	ResourceId          *string `json:"resource_id,omitempty"`
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	IncludeZeroRecord   *bool   `json:"include_zero_record,omitempty"`
	Method              *string `json:"method,omitempty"`
	SubCustomerId       *string `json:"sub_customer_id,omitempty"`
	TradeId             *string `json:"trade_id,omitempty"`
	BillDateBegin       *string `json:"bill_date_begin,omitempty"`
	BillDateEnd         *string `json:"bill_date_end,omitempty"`
}

func (o ListCustomerselfResourceRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerselfResourceRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListCustomerselfResourceRecordsRequest", string(data)}, " ")
}
