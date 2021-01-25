/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type CreateSecurityGroupOption struct {
	// 功能描述：安全组名称 取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）
	Name string `json:"name"`
	// 功能说明：安全组的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description *string `json:"description,omitempty"`
	// 功能说明：企业项目ID。创建安全组时，给安全组绑定企业项目ID。 取值范围：最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。“0”表示默认企业项目。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o CreateSecurityGroupOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupOption struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupOption", string(data)}, " ")
}
