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
type ShowMultiAccountTransferAmountResponse struct {
	// |参数名称：记录条数。| |参数的约束及描述：记录条数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：可拨款余额信息，如果是余额账户，只会有一条记录。具体请参见表 TransferAmountInfoV2。| |参数约束以及描述：可拨款余额信息，如果是余额账户，只会有一条记录。具体请参见表 TransferAmountInfoV2。|
	AmountInfos    *[]TransferAmountInfoV2 `json:"amount_infos,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ShowMultiAccountTransferAmountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowMultiAccountTransferAmountResponse struct{}"
	}

	return strings.Join([]string{"ShowMultiAccountTransferAmountResponse", string(data)}, " ")
}
