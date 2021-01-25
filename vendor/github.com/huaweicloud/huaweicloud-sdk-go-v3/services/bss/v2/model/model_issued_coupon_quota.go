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

type IssuedCouponQuota struct {
	// |参数名称：额度ID。| |参数约束及描述：额度ID。|
	QuotaId *string `json:"quota_id,omitempty"`
	// |参数名称：额度类型：0：代金券额度；| |参数的约束及描述：额度类型：0：代金券额度；|
	QuotaType *int32 `json:"quota_type,omitempty"`
	// |参数名称：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	CreateTime *string `json:"create_time,omitempty"`
	// |参数名称：最后一次更新时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：最后一次更新时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	LastUpdateTime *string `json:"last_update_time,omitempty"`
	// |参数名称：代金券额度的值，精确到小数点后2位。| |参数的约束及描述：代金券额度的值，精确到小数点后2位。|
	QuotaValue float32 `json:"quota_value,omitempty"`
	// |参数名称：状态：0：正常；3：失效（过期失效和人工设置失效）；4：额度调整中（伙伴可以查看该额度，但不能使用该额度发放代金券）。5：冻结6：回收| |参数的约束及描述：状态：0：正常；3：失效（过期失效和人工设置失效）；4：额度调整中（伙伴可以查看该额度，但不能使用该额度发放代金券）。5：冻结6：回收|
	QuotaStatus *int32 `json:"quota_status,omitempty"`
	// |参数名称：剩余的代金券额度，精确到小数点后2位。| |参数的约束及描述：剩余的代金券额度，精确到小数点后2位。|
	Balance float32 `json:"balance,omitempty"`
	// |参数名称：面额单位。1：元。| |参数的约束及描述：面额单位。1：元。|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：币种。当前仅有CNY。| |参数约束及描述：币种。当前仅有CNY。|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：额度上的限制属性| |参数约束以及描述：额度上的限制属性|
	LimitInfos *[]QuotaLimitInfo `json:"limit_infos,omitempty"`
	// |参数名称：二级经销商ID| |参数约束及描述：二级经销商ID|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	// |参数名称：二级经销商账号名称| |参数约束及描述：二级经销商账号名称|
	IndirectPartnerAccountName *string `json:"indirect_partner_account_name,omitempty"`
	// |参数名称：二级经销商名称| |参数约束及描述：二级经销商名称|
	IndirectPartnerName *string `json:"indirect_partner_name,omitempty"`
	// |参数名称：父额度ID，一级经销商用于发给二级经销商额度的额度ID。| |参数约束及描述：父额度ID，一级经销商用于发给二级经销商额度的额度ID。|
	ParentQuotaId *string `json:"parent_quota_id,omitempty"`
}

func (o IssuedCouponQuota) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssuedCouponQuota struct{}"
	}

	return strings.Join([]string{"IssuedCouponQuota", string(data)}, " ")
}
