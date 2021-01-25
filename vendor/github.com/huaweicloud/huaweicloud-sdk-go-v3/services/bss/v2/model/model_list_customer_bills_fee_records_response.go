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

// Response Object
type ListCustomerBillsFeeRecordsResponse struct {
	// |参数名称：总条数，必须大于等于0。| |参数的约束及描述：总条数，必须大于等于0。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：资源费用记录数据。具体请参见表 MonthlyBillRes。| |参数约束以及描述：资源费用记录数据。具体请参见表 MonthlyBillRes。|
	Records *[]MonthlyBillRecord `json:"records,omitempty"`
	// |参数名称：币种。| |参数约束及描述：币种。|
	Currency       *string `json:"currency,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListCustomerBillsFeeRecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerBillsFeeRecordsResponse struct{}"
	}

	return strings.Join([]string{"ListCustomerBillsFeeRecordsResponse", string(data)}, " ")
}
