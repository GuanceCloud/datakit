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
type ListEnterpriseMultiAccountResponse struct {
	// |参数名称：记录条数。| |参数的约束及描述：记录条数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：可回收余额信息，如果是余额账户，只会有一条记录。具体请参见AmountInfo。| |参数约束以及描述：可回收余额信息，如果是余额账户，只会有一条记录。具体请参见AmountInfo。|
	AmountInfos    *[]RetrieveAmountInfoV2 `json:"amount_infos,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ListEnterpriseMultiAccountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnterpriseMultiAccountResponse struct{}"
	}

	return strings.Join([]string{"ListEnterpriseMultiAccountResponse", string(data)}, " ")
}
