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

type QueryCouponQuotasReqExt struct {
	// |参数名称：额度ID列表。| |参数约束以及描述：额度ID列表。|
	QuotaIds *[]string `json:"quota_ids,omitempty"`
	// |参数名称：额度状态列表。| |参数约束以及描述：额度状态列表。|
	QuotaStatusList *[]int32 `json:"quota_status_list,omitempty"`
	// |参数名称：额度类型：0：代金券额度；1：现金券额度。| |参数的约束及描述：额度类型：0：代金券额度；1：现金券额度。|
	QuotaType *int32 `json:"quota_type,omitempty"`
	// |参数名称：创建时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出创建时间大于这个时间的记录。| |参数约束及描述：创建时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出创建时间大于这个时间的记录。|
	CreateTimeBegin *string `json:"create_time_begin,omitempty"`
	// |参数名称：创建时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出创建时间小于这个时间的记录。| |参数约束及描述：创建时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出创建时间小于这个时间的记录。|
	CreateTimeEnd *string `json:"create_time_end,omitempty"`
	// |参数名称：生效时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出生效时间大于这个时间的记录。| |参数约束及描述：生效时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出生效时间大于这个时间的记录。|
	EffectiveTimeBegin *string `json:"effective_time_begin,omitempty"`
	// |参数名称：生效时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出生效时间小于这个时间的记录。| |参数约束及描述：生效时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出生效时间小于这个时间的记录。|
	EffectiveTimeEnd *string `json:"effective_time_end,omitempty"`
	// |参数名称：失效时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出失效时间大于这个时间的记录。| |参数约束及描述：失效时间（开始）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出失效时间大于这个时间的记录。|
	ExpireTimeBegin *string `json:"expire_time_begin,omitempty"`
	// |参数名称：失效时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出失效时间小于这个时间的记录。| |参数约束及描述：失效时间（结束）。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。输入这个条件，会查询出失效时间小于这个时间的记录。|
	ExpireTimeEnd *string `json:"expire_time_end,omitempty"`
	// |参数名称：偏移量，从0开始默认取值为0。| |参数的约束及描述：偏移量，从0开始默认取值为0。|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：每次查询记录数。默认取值为10。| |参数的约束及描述：每次查询记录数。默认取值为10。|
	Limit *int32 `json:"limit,omitempty"`
	// |参数名称：精英服务商（二级经销商）ID，如果要查询二级经销商的额度，需要输入这个参数，否则查询的是一级经销商本人的。| |参数的约束及描述：精英服务商（二级经销商）ID，如果要查询二级经销商的额度，需要输入这个参数，否则查询的是一级经销商本人的。|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o QueryCouponQuotasReqExt) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryCouponQuotasReqExt struct{}"
	}

	return strings.Join([]string{"QueryCouponQuotasReqExt", string(data)}, " ")
}
