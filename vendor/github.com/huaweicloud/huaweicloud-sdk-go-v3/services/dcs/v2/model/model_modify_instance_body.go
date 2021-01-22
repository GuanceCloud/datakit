/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 修改实例信息
type ModifyInstanceBody struct {
	// 实例名称  由英文字符开头，只能由英文字母、数字、中划线和下划线组成。  创建单个实例时，名称长度为4到64位的字符串。批量创建实例时，名称长度为4到56位的字符串，且实例名称格式为“自定义名称-n”，其中n从000开始，依次递增。例如，批量创建两个实例，自定义名称为dcs_demo，则两个实例的名称为dcs_demo-000和dcs_demo-001。
	Name *string `json:"name,omitempty"`
	// 实例的描述信息 长度不超过1024的字符串。 > \\与\"在json报文中属于特殊字符，如果参数值中需要显示\\或者\"字符，请在字符前增加转义字符\\，比如\\\\或者\\\"。
	Description *string `json:"description,omitempty"`
	// '维护时间窗开始时间，为UTC时间，格式为HH:mm:ss。' - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-dcs/ListMaintenanceWindows.html)获取。 - 开始时间必须为22:00:00、02:00:00、06:00:00、10:00:00、14:00:00和18:00:00。 - 该参数不能单独为空，若该值为空，则结束时间也为空。
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// '维护时间窗开始时间，为UTC时间，格式为HH:mm:ss。' - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-dcs/ListMaintenanceWindows.html)获取。 - 结束时间在开始时间基础上加四个小时，即当开始时间为22:00:00时，结束时间为02:00:00。 - 该参数不能单独为空，若该值为空，则开始时间也为空。
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// 安全组ID  可从虚拟私有云服务的控制台界面或者API接口查询得到。  约束：只有Redis 3.0支持
	SecurityGroupId      *string       `json:"security_group_id,omitempty"`
	InstanceBackupPolicy *BackupPolicy `json:"instance_backup_policy,omitempty"`
}

func (o ModifyInstanceBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyInstanceBody struct{}"
	}

	return strings.Join([]string{"ModifyInstanceBody", string(data)}, " ")
}
