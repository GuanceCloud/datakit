/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

//
type SecurityGroupInfo struct {
	// 功能描述：安全组对应的唯一标识 取值范围：带“-”的标准UUID格式
	Id string `json:"id"`
	// 功能说明：安全组名称 取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）
	Name string `json:"name"`
	// 功能说明：安全组的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description string `json:"description"`
	// 功能说明：安全组所属的项目ID
	ProjectId string `json:"project_id"`
	// 功能说明：安全组创建时间 取值范围：UTC时间格式：yyyy-MM-ddTHH:mm:ss
	CreatedAt *sdktime.SdkTime `json:"created_at"`
	// 功能说明：安全组更新时间 取值范围：UTC时间格式：yyyy-MM-ddTHH:mm:ss
	UpdatedAt *sdktime.SdkTime `json:"updated_at"`
	// 功能说明：安全组所属的企业项目ID。 取值范围：最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。“0”表示默认企业项目。
	EnterpriseProjectId string `json:"enterprise_project_id"`
	// 安全组规则
	SecurityGroupRules []SecurityGroupRule `json:"security_group_rules"`
}

func (o SecurityGroupInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroupInfo struct{}"
	}

	return strings.Join([]string{"SecurityGroupInfo", string(data)}, " ")
}
