/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// metadata字段数据结构说明
type MetaDataInfo struct {
	// 用户ID（登录管理控制台，进入我的凭证，即可看到“用户ID”）。
	OpSvcUserid string `json:"op_svc_userid"`
	// 以Windows镜像创建的裸金属服务器Administrator用户的密码，示例：cloud.1234。密码复杂度要求：长度为8-26位。密码至少必须包含大写字母、小写字母、数字和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。密码不能包含用户名或用户名的逆序，不能包含用户名中超过两个连续字符的部分。
	AdminPass *string `json:"admin_pass,omitempty"`
	// 否自带许可，取值“true”或“false”。
	Byol *string `json:"BYOL,omitempty"`
	// 委托的名称。委托是由租户管理员在统一身份认证服务（Identity and Access Management，IAM）上创建的，可以作为其他租户访问此裸金属服务器的临时凭证。 说明:委托获取、更新请参考如下步骤：使用IAM服务提供的查询委托列表，获取有效可用的委托名称。使用更新裸金属服务器元数据接口，更新metadata中agency_name字段为新的委托名称。
	AgencyName *string `json:"agency_name,omitempty"`
}

func (o MetaDataInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MetaDataInfo struct{}"
	}

	return strings.Join([]string{"MetaDataInfo", string(data)}, " ")
}
