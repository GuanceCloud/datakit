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

type AdjustRecordV2 struct {
	// |参数名称：合作伙伴关联的客户的客户ID。| |参数约束及描述：合作伙伴关联的客户的客户ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：合作伙伴关联的客户的客户名。| |参数约束及描述：合作伙伴关联的客户的客户名。|
	CustomerName *string `json:"customer_name,omitempty"`
	// |参数名称：调账类型。0：授信1：回收2：解绑回收| |参数约束及描述：调账类型。0：授信1：回收2：解绑回收|
	OperationType *string `json:"operation_type,omitempty"`
	// |参数名称：调账/回收总额。| |参数的约束及描述：调账/回收总额。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：币种。当前固定为CNY。| |参数约束及描述：币种。当前固定为CNY。|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：使用场景。| |参数约束及描述：使用场景。|
	ApplyScene *string `json:"apply_scene,omitempty"`
	// |参数名称：调账时间。UTC时间，格式为：2016-03-28T14:45:38Z| |参数约束及描述：调账时间。UTC时间，格式为：2016-03-28T14:45:38Z|
	OperationTime *string `json:"operation_time,omitempty"`
	// |参数名称：度量单位。1：元| |参数的约束及描述：度量单位。1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：事务ID，只有在调用3-向客户账户拨款或4-回收客户账户余额接口时，响应消息中返回的该记录存在事务ID“trans_id”字段，这个地方才可能有值。| |参数约束及描述：事务ID，只有在调用3-向客户账户拨款或4-回收客户账户余额接口时，响应消息中返回的该记录存在事务ID“trans_id”字段，这个地方才可能有值。|
	TransId *string `json:"trans_id,omitempty"`
}

func (o AdjustRecordV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AdjustRecordV2 struct{}"
	}

	return strings.Join([]string{"AdjustRecordV2", string(data)}, " ")
}
