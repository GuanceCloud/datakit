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

// Response Object
type CreateInstanceResponse struct {
	// 实例ID。
	Id        *string    `json:"id,omitempty"`
	Datastore *Datastore `json:"datastore,omitempty"`
	// 实例名称，与请求参数相同。
	Name *string `json:"name,omitempty"`
	// 创建时间为本地时间，格式为“yyyy-mm-dd hh:mm:ss”。
	Created *string `json:"created,omitempty"`
	// 实例状态，取值为“creating”。
	Status *string `json:"status,omitempty"`
	// 区域ID，与请求参数相同。
	Region *string `json:"region,omitempty"`
	// 可用区ID，与请求参数相同。
	AvailabilityZone *string `json:"availability_zone,omitempty"`
	// 虚拟私有云ID，与请求参数相同。
	VpcId *string `json:"vpc_id,omitempty"`
	// 子网ID，与请求参数相同。
	SubnetId *string `json:"subnet_id,omitempty"`
	// 实例所属的安全组ID，与请求参数相同。
	SecurityGroupId *string `json:"security_group_id,omitempty"`
	// 磁盘加密的密钥ID，与请求参数相同。
	DiskEncryptionId *string `json:"disk_encryption_id,omitempty"`
	// 实例类型，与请求参数相同。
	Mode *string `json:"mode,omitempty"`
	// 实例规格详情，与请求参数相同。
	Flavor         *[]CreateInstanceFlavorOption `json:"flavor,omitempty"`
	BackupStrategy *BackupStrategy               `json:"backup_strategy,omitempty"`
	// 企业项目ID。取值为“0”，表示为default企业项目。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// SSL开关选项，与请求参数相同。
	SslOption *string `json:"ssl_option,omitempty"`
	// 专属存储池ID。
	DssPoolId *string `json:"dss_pool_id,omitempty"`
	// 创建实例的工作流ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceResponse struct{}"
	}

	return strings.Join([]string{"CreateInstanceResponse", string(data)}, " ")
}
