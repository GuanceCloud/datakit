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

type CreateSubCustomerReqV2 struct {
	// |参数名称：子账号挂载的组织单元，填写组织单元的Party ID，通过查询企业组织结构接口的响应获得。| |参数约束及描述：子账号挂载的组织单元，填写组织单元的Party ID，通过查询企业组织结构接口的响应获得。|
	PartyId string `json:"party_id"`
	// |参数名称：企业子账号的显示名称不限制特殊字符。| |参数约束及描述：企业子账号的显示名称不限制特殊字符。|
	DisplayName *string `json:"display_name,omitempty"`
	// |参数名称：子账号关联类型：1：同一法人。注：关联类型目前只能是同一法人。| |参数的约束及描述：子账号关联类型：1：同一法人。注：关联类型目前只能是同一法人。|
	SubCustomerAssociationType *int32 `json:"sub_customer_association_type,omitempty"`
	// |参数名称：申请的权限列表。支持的权限项参见表 权限项定义列表| |参数约束以及描述：申请的权限列表。支持的权限项参见表 权限项定义列表|
	PermissionIds  *[]string      `json:"permission_ids,omitempty"`
	NewSubCustomer *NewCustomerV2 `json:"new_sub_customer"`
}

func (o CreateSubCustomerReqV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubCustomerReqV2 struct{}"
	}

	return strings.Join([]string{"CreateSubCustomerReqV2", string(data)}, " ")
}
