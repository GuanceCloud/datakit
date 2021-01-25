/*
 * DDS
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
type CreateInstanceRequestBody struct {
	// 实例名称。用于表示实例的名称，用于表示实例的名称，同一租户下，同类型的实例名唯一。 取值范围：长度为4~64位，必须以字母开头（A~Z或a~z），区分大小写，可以包含字母、数字（0~9）、中划线（-）或者下划线（_），不能包含其他特殊字符。
	Name      string     `json:"name"`
	Datastore *Datastore `json:"datastore"`
	// 区域ID，恢复到新实例时不可选。
	Region string `json:"region"`
	// 可用区ID。
	AvailabilityZone string `json:"availability_zone"`
	// 虚拟私有云ID。获取方法请参见《虚拟私有云API参考》中“VPC”的内容。 取值：非空，字符长度校验，严格UUID正则校验。
	VpcId string `json:"vpc_id"`
	// 子网ID。获取方法请参见《虚拟私有云API参考》中“子网”的内容。
	SubnetId string `json:"subnet_id"`
	// 指定实例所属的安全组ID。 获取方法请参见《虚拟私有云API参考》中“安全组”的内容。
	SecurityGroupId string `json:"security_group_id"`
	// 数据库密码。 取值范围：长度为8~32位，必须是大写字母（A~Z）、小写字母（a~z）、数字（0~9）、特殊字符~!@#%^*-_=+?的组合。 建议您输入高强度密码，以提高安全性，防止出现密码被暴力破解等安全风险。
	Password *string `json:"password,omitempty"`
	// 磁盘加密时的密钥ID，严格UUID正则校验。 不传该参数时，表示不进行磁盘加密。
	DiskEncryptionId *string `json:"disk_encryption_id,omitempty"`
	// 实例类型。支持集群、副本集、以及单节点。 取值   - Sharding   - ReplicaSet   - Single
	Mode string `json:"mode"`
	// 实例规格详情。
	Flavor         []CreateInstanceFlavorOption `json:"flavor"`
	BackupStrategy *BackupStrategy              `json:"backup_strategy,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// SSL开关选项。 取值： - 取“0”，表示DDS实例默认不启用SSL连接。 - 取“1”，表示DDS实例默认启用SSL连接。 - 不传该参数时，默认启用SSL连接。
	SslOption *string `json:"ssl_option,omitempty"`
	// 创建新实例所在专属存储池ID，仅专属云创建实例时有效。
	DssPoolId *string `json:"dss_pool_id,omitempty"`
	// 创建新实例设置云服务器组关联的策略名称列表，仅专属云创建实例时有效。 取值    - 取“anti-affinity”，表示DDS实例开启反亲和部署，反亲和部署是出于高可用性考虑，将您的Primary、Secondary和Hidden节点分别创建在不同的物理机上。当前仅支持该值，不传该值默认不开启反亲和部署。
	ServerGroupPolicies *[]string     `json:"server_group_policies,omitempty"`
	RestorePoint        *RestorePoint `json:"restore_point,omitempty"`
}

func (o CreateInstanceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceRequestBody struct{}"
	}

	return strings.Join([]string{"CreateInstanceRequestBody", string(data)}, " ")
}
