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
type ListCustomerBillsFeeRecordsRequest struct {
	XLanguage           *string `json:"X-Language,omitempty"`
	BillCycle           string  `json:"bill_cycle"`
	ProviderType        *int32  `json:"provider_type,omitempty"`
	ServiceTypeCode     *string `json:"service_type_code,omitempty"`
	ResourceTypeCode    *string `json:"resource_type_code,omitempty"`
	RegionCode          *string `json:"region_code,omitempty"`
	ChargingMode        *int32  `json:"charging_mode,omitempty"`
	BillType            *int32  `json:"bill_type,omitempty"`
	TradeId             *string `json:"trade_id,omitempty"`
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	IncludeZeroRecord   *bool   `json:"include_zero_record,omitempty"`
	Status              *int32  `json:"status,omitempty"`
	Method              *string `json:"method,omitempty"`
	SubCustomerId       *string `json:"sub_customer_id,omitempty"`
	Offset              *int32  `json:"offset,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
}

func (o ListCustomerBillsFeeRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerBillsFeeRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListCustomerBillsFeeRecordsRequest", string(data)}, " ")
}
