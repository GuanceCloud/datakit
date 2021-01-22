/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 资源配额实例详情
type QuotaShowResp struct {
	// 功能说明：根据type过滤查询指定类型的配额 取值范围：vpc，subnet，securityGroup，securityGroupRule，publicIp，vpn，vpngw，vpcPeer，firewall，shareBandwidth，shareBandwidthIP
	Type *string `json:"type,omitempty"`
	// 功能说明：已创建的资源个数 取值范围：0~quota数
	Used *int32 `json:"used,omitempty"`
	// 功能说明：资源的最大配额数 取值范围：各类型资源默认配额数~Integer最大值 约束：资源的默认配额数可以修改，而且配额需要提前在底层配置，参考默认配置为：vpc默认5，子网默认100，安全组默认100，安全组规则默认5000，弹性公网IP默认10，vpn默认5，vpngw默认2，vpcPeer默认50，firewall默认200，shareBandwidth默认5，shareBandwidthIP默认20
	Quota *int32 `json:"quota,omitempty"`
	// 允许修改的配额最小值
	Min *int32 `json:"min,omitempty"`
}

func (o QuotaShowResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuotaShowResp struct{}"
	}

	return strings.Join([]string{"QuotaShowResp", string(data)}, " ")
}
