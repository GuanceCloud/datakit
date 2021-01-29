/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 实例信息。
type InstanceRequest struct {
	// 实例名称。 用于表示实例的名称，同一租户下，同类型的实例名可重名，其中，SQL Server实例名唯一。 取值范围：4~64个字符之间，必须以字母开头，区分大小写，可以包含字母、数字、中划线或者下划线，不能包含其他的特殊字符。
	Name      string     `json:"name"`
	Datastore *Datastore `json:"datastore"`
	Ha        *Ha        `json:"ha,omitempty"`
	// 参数组ID。
	ConfigurationId *string `json:"configuration_id,omitempty"`
	// 数据库端口信息。  - MySQL数据库端口设置范围为1024～65535（其中12017和33071被RDS系统占用不可设置）。 - PostgreSQL数据库端口修改范围为2100～9500。 - Microsoft SQL Server实例的端口设置范围为1433和2100~9500（其中5355和5985不可设置。对于2017 EE、2017 SE、2017 Web版，5050、5353和5986不可设置。  当不传该参数时，默认端口如下：  - MySQL默认3306。 - PostgreSQL默认5432。 - Microsoft SQL Server默认1433。
	Port *string `json:"port,omitempty"`
	// 数据库密码。创建只读实例时不可选，其它场景必选。  取值范围：  非空，由大小写字母、数字和特殊符号~!@#%^*-_=+?组成，长度8~32个字符。  建议您输入高强度密码，以提高安全性，防止出现密码被暴力破解等安全风险。
	Password       *string         `json:"password,omitempty"`
	BackupStrategy *BackupStrategy `json:"backup_strategy,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 用于磁盘加密的密钥ID。
	DiskEncryptionId *string `json:"disk_encryption_id,omitempty"`
	// 规格码。
	FlavorRef string  `json:"flavor_ref"`
	Volume    *Volume `json:"volume"`
	// 区域ID。创建主实例时必选，其它场景不可选。 取值参见[地区和终端节点](https://developer.huaweicloud.com/endpoint)。
	Region string `json:"region"`
	// 可用区ID。对于数据库实例类型不是单机的实例，需要分别为实例所有节点指定可用区，并用逗号隔开。 取值参见[地区和终端节点](https://developer.huaweicloud.com/endpoint)。
	AvailabilityZone string `json:"availability_zone"`
	// 虚拟私有云ID。创建只读实例时不可选（只读实例的网络属性默认和主实例相同），其它场景必选。
	VpcId string `json:"vpc_id"`
	// 子网ID。创建只读实例时不可选（只读实例的网络属性默认和主实例相同），其它场景必选。
	SubnetId string `json:"subnet_id"`
	// 指定实例的内网IP
	DataVip *string `json:"data_vip,omitempty"`
	// 安全组ID。创建只读实例时不可选（只读实例的网络属性默认和主实例相同），其它场景必选。
	SecurityGroupId string      `json:"security_group_id"`
	ChargeInfo      *ChargeInfo `json:"charge_info,omitempty"`
	// 时区。  - 不选择时，各个引擎时区如下：   - MySQL国内站、国际站默认为UTC时间。   - PostgreSQL国内站、国际站默认为UTC时间。   - Microsoft SQL Server国内站默认为UTC+08:00，国际站默认为UTC时间。 - 选择填写时，取值范围为UTC-12:00~UTC+12:00，且只支持整段时间，如UTC+08:00，不支持UTC+08:30。
	TimeZone *string `json:"time_zone,omitempty"`
	// 专属存储池ID。
	DsspoolId *string `json:"dsspool_id,omitempty"`
	// 只读实例的主实例ID。创建只读实例时必选，其它场景不可选。
	ReplicaOfId  *string       `json:"replica_of_id,omitempty"`
	RestorePoint *RestorePoint `json:"restore_point,omitempty"`
	// 仅限Microsoft SQL Server实例使用。取值范围：根据查询SQL Server可用字符集的字符集查询列表查询可设置的字符集。
	Collation *string `json:"collation,omitempty"`
	// 标签列表。单个实例总标签数上限10个。
	Tags *[]InstanceRequestTags `json:"tags,omitempty"`
}

func (o InstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceRequest struct{}"
	}

	return strings.Join([]string{"InstanceRequest", string(data)}, " ")
}
