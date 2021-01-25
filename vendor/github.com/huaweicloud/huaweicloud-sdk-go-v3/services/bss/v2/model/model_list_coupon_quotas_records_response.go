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
type ListCouponQuotasRecordsResponse struct {
	// |参数名称：查询总数。| |参数的约束及描述：查询总数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：记录列表。具体请参见表 QuotaRecord。| |参数约束以及描述：记录列表。具体请参见表 QuotaRecord。|
	Records        *[]QuotaRecord `json:"records,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListCouponQuotasRecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCouponQuotasRecordsResponse struct{}"
	}

	return strings.Join([]string{"ListCouponQuotasRecordsResponse", string(data)}, " ")
}
