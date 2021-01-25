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

type RetrieveEnterpriseMultiAccountReq struct {
	// |参数名称：企业子账号的客户ID。| |参数约束及描述：企业子账号的客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：总回收金额。金额单位为货币标准单位，如人民币则单位为元。| |参数约束及描述：总回收金额。金额单位为货币标准单位，如人民币则单位为元。|
	Amount string `json:"amount"`
	// |参数名称：交易序列号，用于防止重复提交。如果接口调用方不传，则企业管理微服务后台生成。如果接口调用方传入，则请采用UUID保证全局唯一。| |参数约束及描述：交易序列号，用于防止重复提交。如果接口调用方不传，则企业管理微服务后台生成。如果接口调用方传入，则请采用UUID保证全局唯一。|
	TransId *string `json:"trans_id,omitempty"`
	// |参数名称：账户类型：BALANCE_TYPE_DEBIT：余额账户（默认）；BALANCE_TYPE_CREDIT：信用账户。| |参数约束及描述：账户类型：BALANCE_TYPE_DEBIT：余额账户（默认）；BALANCE_TYPE_CREDIT：信用账户。|
	BalanceType *string `json:"balance_type,omitempty"`
	// |参数名称：账户到期时间，UTC时间，格式为：2016-03-28T14:45:38Z。暂只对信用账户有效，用于限制针对有效期到期时间等于该时间的信用账户余额进行回收。如果查询信用账户可回收余额的查询结果没有失效时间，表示永久有效，对于这种账本的回收的时候不用填写| |参数约束及描述：账户到期时间，UTC时间，格式为：2016-03-28T14:45:38Z。暂只对信用账户有效，用于限制针对有效期到期时间等于该时间的信用账户余额进行回收。如果查询信用账户可回收余额的查询结果没有失效时间，表示永久有效，对于这种账本的回收的时候不用填写|
	ExpireTime *string `json:"expire_time,omitempty"`
}

func (o RetrieveEnterpriseMultiAccountReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RetrieveEnterpriseMultiAccountReq struct{}"
	}

	return strings.Join([]string{"RetrieveEnterpriseMultiAccountReq", string(data)}, " ")
}
