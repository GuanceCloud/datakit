/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowResourceByIdResponse struct {
	// 资源id
	Id *string `json:"id,omitempty"`
	// 资源名称
	Name *string `json:"name,omitempty"`
	// 云服务名称
	Provider *string `json:"provider,omitempty"`
	// 资源类型
	Type *string `json:"type,omitempty"`
	// region id
	RegionId *string `json:"region_id,omitempty"`
	// Openstack中的project id
	ProjectId *string `json:"project_id,omitempty"`
	// Openstack中的project名称
	ProjectName *string `json:"project_name,omitempty"`
	// 企业项目id
	EpId *string `json:"ep_id,omitempty"`
	// 企业项目名称
	EpName *string `json:"ep_name,omitempty"`
	// 资源详情校验码
	Checksum *string `json:"checksum,omitempty"`
	// 资源创建时间
	Created *string `json:"created,omitempty"`
	// 资源更新时间
	Updated *string `json:"updated,omitempty"`
	// 资源操作状态
	ProvisioningState *string `json:"provisioning_state,omitempty"`
	// 资源Tag
	Tags map[string]string `json:"tags,omitempty"`
	// 资源详细属性
	Properties     map[string]interface{} `json:"properties,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o ShowResourceByIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceByIdResponse struct{}"
	}

	return strings.Join([]string{"ShowResourceByIdResponse", string(data)}, " ")
}
