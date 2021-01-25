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

type DemandProductInfo struct {
	// |参数名称：ID标识| |参数约束及描述：同一次询价中不能重复，用于标识返回询价结果和请求的映射关系|
	Id string `json:"id"`
	// |参数名称：用户购买云服务产品的云服务类型| |参数约束及描述：例如EC2，云服务类型为hws.service.type.ec2|
	CloudServiceType string `json:"cloud_service_type"`
	// |参数名称：用户购买云服务产品的资源类型| |参数约束及描述：例如EC2中的VM，资源类型为hws.resource.type.vm。ResourceType是CloudServiceType中的一种资源，CloudServiceType由多种ResourceType组合提供|
	ResourceType string `json:"resource_type"`
	// |参数名称：用户购买云服务产品的资源规格| |参数约束及描述：例如VM的小型规格，资源规格为m1.tiny|
	ResourceSpec string `json:"resource_spec"`
	// |参数名称：云服务区编码| |参数约束及描述：云服务区编码|
	Region string `json:"region"`
	// |参数名称：可用区标识| |参数约束及描述：可用区标识|
	AvailableZone *string `json:"available_zone,omitempty"`
	// |参数名称：资源容量大小| |参数约束及描述：例如购买的卷大小或带宽大小，只有线性产品才有这个字段|
	ResourceSize *int32 `json:"resource_size,omitempty"`
	// |参数名称：资源容量度量标识| |参数约束及描述：枚举值如下：15：Mbps（购买带宽时使用）17：GB（购买云硬盘时使用）14：个只有线性产品才有这个字段|
	SizeMeasureId *int32 `json:"size_measure_id,omitempty"`
	// |参数名称：使用量因子编码| |参数约束及描述：云服务器：Duration云硬盘：Duration弹性IP：Duration带宽：Duration或upflow市场镜像：Duration具体每种云服务使用什么样的计费因子，需要找具体云服务确认，全集请参考|
	UsageFactor string `json:"usage_factor"`
	// |参数名称：使用量值| |参数约束及描述：例如按小时询价，使用量值为1，使用量单位为小时|
	UsageValue float32 `json:"usage_value"`
	// |参数名称：使用量单位标识| |参数约束及描述：例如按小时询价，使用量值为1，使用量单位为小时，枚举值如下：4：小时全量枚举如下：0：天（时长）；1：元（货币）；2：角（货币）；3：分（货币）；4：小时（时长）；5：分钟（时长）；6：秒（时长）；7：EB（流量）；8：PB（流量）；9：TB（流量）；10：GB（流量）；11：MB（流量）；12：KB（流量）；13：Byte（流量）；14：个(次)（数量）；15：Mbps（流量）；16：Byte（容量）；17：GB（容量）；18：KLOC（行数）；19：年（周期）；20：月（周期）；21：MB（容量）；22：赫兹（频率）；23：核（数量）；24：天（周期）；25：小时（周期）；30：个数（个数）；31：千次（数量）；32：百万次（数量）；33：十亿次（数量）；34：bps（带宽速率）；35：kbps（带宽速率）；36：Mbps（带宽速率）；37：Gbps（带宽速率）；38：Tbps（带宽速率）；39：GB-秒（容量时长）；40：次（数量）；41：个（数量）；42：千个（数量）；43：张（数量）；44：千张（数量）；45：每秒查询率（查询速率）；46：人/天（数量）；47：TB（容量）；48：PB（容量）。具体某个云服务应该使用什么单位，需要和云服务确认。|
	UsageMeasureId int32 `json:"usage_measure_id"`
	// |参数名称：订购数量| |参数约束及描述：订购数量,有值时不能小于0，默认为1|
	SubscriptionNum int32 `json:"subscription_num"`
}

func (o DemandProductInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DemandProductInfo struct{}"
	}

	return strings.Join([]string{"DemandProductInfo", string(data)}, " ")
}
