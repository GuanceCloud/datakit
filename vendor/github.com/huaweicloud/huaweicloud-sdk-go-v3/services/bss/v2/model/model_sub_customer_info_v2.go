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

type SubCustomerInfoV2 struct {
	// |参数名称：企业子账号的客户ID。| |参数约束及描述：企业子账号的客户ID。|
	Id *string `json:"id,omitempty"`
	// |参数名称：企业子账号的用户名。| |参数约束及描述：企业子账号的用户名。|
	Name *string `json:"name,omitempty"`
	// |参数名称：企业子账号的显示名称。不限制特殊字符。| |参数约束及描述：企业子账号的显示名称。不限制特殊字符。|
	DisplayName *string `json:"display_name,omitempty"`
	// |参数名称：子账号状态：1：正常；2：创建中；3：关闭中；4：已关闭；101：子账号注册中；102：子账号待激活。| |参数的约束及描述：子账号状态：1：正常；2：创建中；3：关闭中；4：已关闭；101：子账号注册中；102：子账号待激活。|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：子账号归属的组织单元ID| |参数约束及描述：子账号归属的组织单元ID|
	OrgId *string `json:"org_id,omitempty"`
	// |参数名称：子账号归属的组织单元名称注：当子账号归属的组织是企业组织根节点时，本属性可能为空。| |参数约束及描述：子账号归属的组织单元名称注：当子账号归属的组织是企业组织根节点时，本属性可能为空。|
	OrgName *string `json:"org_name,omitempty"`
}

func (o SubCustomerInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SubCustomerInfoV2 struct{}"
	}

	return strings.Join([]string{"SubCustomerInfoV2", string(data)}, " ")
}
