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

type PackageUsageInfo struct {
	// |参数名称：订购实例ID| |参数的约束及描述：订购实例ID|
	OrderInstanceId *string `json:"order_instance_id,omitempty"`
	// |参数名称：资源类型名称| |参数的约束及描述：资源类型名称|
	ResourceTypeName *string `json:"resource_type_name,omitempty"`
	// |参数名称：重用模式| |参数的约束及描述：重用模式: 1：可重用2：不可重用|
	QuotaReuseMode *int32 `json:"quota_reuse_mode,omitempty"`
	// |参数名称：重用周期| |参数的约束及描述：重用周期，只有quotaReuseMode为可重用，该字段才有意义.1：小时2：天3：周4：月5：年|
	QuotaReuseCycle *int32 `json:"quota_reuse_cycle,omitempty"`
	// |参数名称：重用周期类别| |参数的约束及描述：重置周期类别，只有quotaReuseMode为可重用，该字段才有意义1：按自然周期重置2：按订购周期重置|
	QuotaReuseCycleType *int32 `json:"quota_reuse_cycle_type,omitempty"`
	// |参数名称：开始时间，格式UTC| |参数的约束及描述：1）如果quotaReuseMode为可重用，则此时间为当前时间所在的重用周期的开始时间2）如果quotaReuseMode为不可重用，则此时间为订购实例的生效时间，|
	StartTime *string `json:"start_time,omitempty"`
	// |参数名称：结束时间，格式UTC| |参数的约束及描述：1）如果quotaReuseMode为可重用，则此时间为当前时间所在的重用周期的结束时间2）如果quotaReuseMode为不可重用，则此时间为订购实例的失效时间|
	EndTime *string `json:"end_time,omitempty"`
	// |参数名称：套餐包内资源剩余量| |参数的约束及描述：套餐包内资源剩余量|
	Balance float32 `json:"balance,omitempty"`
	// |参数名称：套餐包的资源总量| |参数的约束及描述：套餐包的资源总量|
	Total float32 `json:"total,omitempty"`
	// |参数名称：套餐包资源的度量单位名称| |参数的约束及描述：套餐包资源的度量单位名称|
	MeasurementName *string `json:"measurement_name,omitempty"`
	// |参数名称：区域编码| |参数的约束及描述：区域编码|
	RegionCode *string `json:"region_code,omitempty"`
}

func (o PackageUsageInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PackageUsageInfo struct{}"
	}

	return strings.Join([]string{"PackageUsageInfo", string(data)}, " ")
}
