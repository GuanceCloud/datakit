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
type ListCustomerselfResourceRecordDetailsResponse struct {
	// |参数名称：资源费用记录数据| |参数的约束及描述：该参数非必填|
	MonthlyRecords *[]MonthlyBillRes `json:"monthly_records,omitempty"`
	// |参数名称：结果集数量| |参数的约束及描述：该参数非必填，且只允许数字，只有成功才返回这个参数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：货币单位代码| |参数的约束及描述：该参数非必填，最大长度3，CNY：人民币；USD：美元|
	Currency       *string `json:"currency,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListCustomerselfResourceRecordDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerselfResourceRecordDetailsResponse struct{}"
	}

	return strings.Join([]string{"ListCustomerselfResourceRecordDetailsResponse", string(data)}, " ")
}
