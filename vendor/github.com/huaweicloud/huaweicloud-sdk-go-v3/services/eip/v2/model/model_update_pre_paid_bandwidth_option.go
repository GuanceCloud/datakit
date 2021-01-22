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

// 更新带宽请求中的带宽信息
type UpdatePrePaidBandwidthOption struct {
	// 功能说明：带宽名称  取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点），为空表示不修改名称  约束：和参数size必须有一个参数有值
	Name *string `json:"name,omitempty"`
	// 功能说明：带宽大小，包周期的带宽只能改大  取值范围：默认1Mbit/s～2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示），不带此参数时表示不修改大小。  约束：和参数name必须有一个参数有值。  注意：调整带宽时的最小单位会根据带宽范围不同存在差异。  小于等于300Mbit/s：默认最小单位为1Mbit/s。  300Mbit/s~1000Mbit/s：默认最小单位为50Mbit/s。  大于1000Mbit/s：默认最小单位为500Mbit/s。
	Size *int32 `json:"size,omitempty"`
}

func (o UpdatePrePaidBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePrePaidBandwidthOption struct{}"
	}

	return strings.Join([]string{"UpdatePrePaidBandwidthOption", string(data)}, " ")
}
