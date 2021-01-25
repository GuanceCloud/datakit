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

// Response Object
type ListEnterpriseOrganizationsResponse struct {
	// |参数名称：根节点ID，如果请求有parent_id，则该参数无值。| |参数约束及描述：根节点ID，如果请求有parent_id，则该参数无值。|
	RootId *string `json:"root_id,omitempty"`
	// |参数名称：根节点名称，如果请求有parent_id，则该参数无值。注：组织根节点没有设置组织名称时，可能为空。| |参数约束及描述：根节点名称，如果请求有parent_id，则该参数无值。注：组织根节点没有设置组织名称时，可能为空。|
	RootName *string `json:"root_name,omitempty"`
	// |参数名称：企业管理子Party节点列表。注：每一层的节点列表需要按relation_type升序排序。| |参数约束以及描述：企业管理子Party节点列表。注：每一层的节点列表需要按relation_type升序排序。|
	ChildNodes     *[]EmChildNodeV2 `json:"child_nodes,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListEnterpriseOrganizationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnterpriseOrganizationsResponse struct{}"
	}

	return strings.Join([]string{"ListEnterpriseOrganizationsResponse", string(data)}, " ")
}
