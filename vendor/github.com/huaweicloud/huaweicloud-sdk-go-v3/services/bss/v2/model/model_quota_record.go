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

type QuotaRecord struct {
	// |参数名称：记录ID| |参数约束及描述：记录ID|
	Id *string `json:"id,omitempty"`
	// |参数名称：操作员额账号名称| |参数约束及描述：操作员额账号名称|
	Operator *string `json:"operator,omitempty"`
	// |参数名称：操作类型10：发放额度11：回收额度| |参数约束及描述：操作类型10：发放额度11：回收额度|
	OperationType *string `json:"operation_type,omitempty"`
	// |参数名称：额度ID，这里指的是一级经销商发给二级经销商额度时，产生的二级经销商的额度ID，或者从二级经销商回收的时候，二级经销商的额度ID| |参数约束及描述：额度ID，这里指的是一级经销商发给二级经销商额度时，产生的二级经销商的额度ID，或者从二级经销商回收的时候，二级经销商的额度ID|
	QuotaId *string `json:"quota_id,omitempty"`
	// |参数名称：父额度ID，这里指的是一级经销商发给二级经销商额度时，一级经销商的额度ID，或者从二级经销商回收的时候，回收到的一级经销商的额度ID| |参数约束及描述：父额度ID，这里指的是一级经销商发给二级经销商额度时，一级经销商的额度ID，或者从二级经销商回收的时候，回收到的一级经销商的额度ID|
	ParentQuotaId *string `json:"parent_quota_id,omitempty"`
	// |参数名称：发放回收的金额，小数点后2位，单位元| |参数的约束及描述：发放回收的金额，小数点后2位，单位元|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：操作时间，UTC时间，UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：操作时间，UTC时间，UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	OperationTime *string `json:"operation_time,omitempty"`
	// |参数名称：操作结果0：成功-1：失败| |参数约束及描述：操作结果0：成功-1：失败|
	Result *string `json:"result,omitempty"`
	// |参数名称：二级经销商的管理员账号名| |参数约束及描述：二级经销商的管理员账号名|
	IndirectPartnerAccountName *string `json:"indirect_partner_account_name,omitempty"`
	// |参数名称：二级经销商ID| |参数约束及描述：二级经销商ID|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	// |参数名称：二级经销商的公司名称| |参数约束及描述：二级经销商的公司名称|
	IndirectPartnerName *string `json:"indirect_partner_name,omitempty"`
	// |参数名称：备注| |参数约束及描述：备注|
	Remark *string `json:"remark,omitempty"`
}

func (o QuotaRecord) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuotaRecord struct{}"
	}

	return strings.Join([]string{"QuotaRecord", string(data)}, " ")
}
