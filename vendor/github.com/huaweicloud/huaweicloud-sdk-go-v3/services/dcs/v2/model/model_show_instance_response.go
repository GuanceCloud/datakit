/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowInstanceResponse struct {
	// VPC的名称。
	VpcName *string `json:"vpc_name,omitempty"`
	// 付费模式，0表示按需计费，1表示包年/包月计费。
	ChargingMode *int32 `json:"charging_mode,omitempty"`
	// VPC ID
	VpcId *string `json:"vpc_id,omitempty"`
	// 用户名。
	UserName *string `json:"user_name,omitempty"`
	// 完成创建时间。格式为：2017-03-31T12:24:46.297Z
	CreatedAt *string `json:"created_at,omitempty"`
	// 实例描述。
	Description *string `json:"description,omitempty"`
	// 安全组ID。
	SecurityGroupId *string `json:"security_group_id,omitempty"`
	// 租户安全组名称。
	SecurityGroupName *string `json:"security_group_name,omitempty"`
	// 总内存，单位：MB。
	MaxMemory *int32 `json:"max_memory,omitempty"`
	// 已使用的内存，单位：MB。
	UsedMemory *int32 `json:"used_memory,omitempty"`
	// 缓存实例的容量（G Byte）。
	Capacity *int32 `json:"capacity,omitempty"`
	// 单机小规格的缓存容量。
	CapacityMinor *string `json:"capacity_minor,omitempty"`
	// 维护时间窗开始时间，为UTC时间，格式为HH:mm:ss
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// 维护时间窗结束时间，为UTC时间，格式为HH:mm:ss
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// 缓存实例的引擎类型。
	Engine *string `json:"engine,omitempty"`
	// 是否允许免密码访问缓存实例。 - true：该实例无需密码即可访问。 - false：该实例必须通过密码认证才能访问。
	NoPasswordAccess *string `json:"no_password_access,omitempty"`
	// 连接缓存实例的IP地址。如果是集群实例，返回多个IP地址，使用逗号分隔。如：192.168.0.1，192.168.0.2。
	Ip                   *string       `json:"ip,omitempty"`
	InstanceBackupPolicy *BackupPolicy `json:"instance_backup_policy,omitempty"`
	// 实例所在的可用区。返回“可用区Code”
	AzCodes *[]string `json:"az_codes,omitempty"`
	// 通过密码认证访问缓存实例的认证用户名。
	AccessUser *string `json:"access_user,omitempty"`
	// 实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 缓存的端口。
	Port *int32 `json:"port,omitempty"`
	// 用户id。
	UserId *string `json:"user_id,omitempty"`
	// 实例名称。
	Name *string `json:"name,omitempty"`
	// 产品规格编码
	SpecCode *string `json:"spec_code,omitempty"`
	// 子网ID。
	SubnetId *string `json:"subnet_id,omitempty"`
	// 子网名称。
	SubnetName *string `json:"subnet_name,omitempty"`
	// 子网网段。
	SubnetCidr *string `json:"subnet_cidr,omitempty"`
	// 缓存版本。
	EngineVersion *string `json:"engine_version,omitempty"`
	// 订单ID。
	OrderId *string `json:"order_id,omitempty"`
	// 缓存实例的状态。详细状态说明见[缓存实例状态说明](https://support.huaweicloud.com/api-dcs/dcs-api-0312047.html)
	Status *string `json:"status,omitempty"`
	// 实例的域名。
	DomainName *string `json:"domain_name,omitempty"`
	// Redis缓存实例是否开启公网访问功能。 - true：开启 - false：不开启
	EnablePublicip *bool `json:"enable_publicip,omitempty"`
	// Redis缓存实例绑定的弹性IP地址的id。 如果未开启公网访问功能，该字段值为null。
	PublicipId *string `json:"publicip_id,omitempty"`
	// Redis缓存实例绑定的弹性IP地址。 如果未开启公网访问功能，该字段值为null。
	PublicipAddress *string `json:"publicip_address,omitempty"`
	// Redis缓存实例开启公网访问功能时，是否选择支持ssl。 - true：开启 - false：不开启
	EnableSsl *bool `json:"enable_ssl,omitempty"`
	// 实例是否存在升级任务。 - true：存在 - false：不存在
	ServiceUpgrade *bool `json:"service_upgrade,omitempty"`
	// 升级任务的ID。 - 当service_upgrade为true时，为升级任务的ID。 - 当service_upgrade为false时，该参数为空。
	ServiceTaskId *string `json:"service_task_id,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 集群实例的后端服务地址。
	BackendAddrs   *string `json:"backend_addrs,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceResponse struct{}"
	}

	return strings.Join([]string{"ShowInstanceResponse", string(data)}, " ")
}
