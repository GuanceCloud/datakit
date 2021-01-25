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
type QueryInstanceResponse struct {
	// 实例ID。
	Id string `json:"id"`
	// 实例名称。
	Name string `json:"name"`
	// 实例状态。 取值： - normal，表示实例正常。 - abnormal，表示实例异常。 - creating，表示实例创建中。 - frozen，表示实例被冻结。 - data_disk_full，表示实例磁盘已满。 - createfail，表示实例创建失败。 - enlargefail，表示实例扩容节点个数失败。
	Status string `json:"status"`
	// 数据库端口号。文档数据库实例支持的端口号范围为2100～9500。
	Port int32 `json:"port"`
	// 实例类型。与请求参数相同。
	Mode string `json:"mode"`
	// 实例所在区域。
	Region    string         `json:"region"`
	Datastore *DatastoreItem `json:"datastore"`
	// 存储引擎。取值为“wiredTiger”。
	Engine string `json:"engine"`
	// 实例创建时间。
	Created string `json:"created"`
	// 实例操作最新变更的时间。
	Updated string `json:"updated"`
	// 默认用户名。取值为“rwuser”。
	DbUserName string `json:"db_user_name"`
	// 是否开启SSL安全连接。 - 取值为“1”，表示开启。 - 取值为“0”，表示不开启。
	Ssl int32 `json:"ssl"`
	// 虚拟私有云ID。
	VpcId string `json:"vpc_id"`
	// 子网ID。
	SubnetId string `json:"subnet_id"`
	// 安全组ID。
	SecurityGroupId string                         `json:"security_group_id"`
	BackupStrategy  *BackupStrategyForItemResponse `json:"backup_strategy"`
	// 计费方式。 - 取值为“0”，表示按需计费。 - 取值为“1”，表示包年/包月计费。
	PayMode *string `json:"pay_mode,omitempty"`
	// 系统可维护时间窗。
	MaintenanceWindow string `json:"maintenance_window"`
	// 组信息。
	Groups []GroupResponseItem `json:"groups"`
	// 磁盘加密的密钥ID。
	DiskEncryptionId string `json:"disk_encryption_id"`
	// 企业项目ID。取值为“0”，表示为default企业项目。
	EnterpriseProjectId string `json:"enterprise_project_id"`
	// 时区。
	TimeZone string `json:"time_zone"`
	// 专属存储池ID。
	DssPoolId *string `json:"dss_pool_id,omitempty"`
	// 实例正在执行的动作。
	Actions []string `json:"actions"`
}

func (o QueryInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryInstanceResponse struct{}"
	}

	return strings.Join([]string{"QueryInstanceResponse", string(data)}, " ")
}
