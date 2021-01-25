/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type UpdateInstanceReq struct {
	// 实例名称。  由英文字符开头，只能由英文字母、数字、中划线组成，长度为4~64的字符。
	Name *string `json:"name,omitempty"`
	// 实例的描述信息。  长度不超过1024的字符串。  > \\与\"在json报文中属于特殊字符，如果参数值中需要显示\\或者\"字符，请在字符前增加转义字符\\，比如\\\\或者\\\"。
	Description *string `json:"description,omitempty"`
	// 维护时间窗开始时间，格式为HH:mm:ss。   - 维护时间窗开始和结束时间必须为指定的时间段，可参考查[询维护时间窗时间段](https://support.huaweicloud.com/api-kafka/ShowMaintainWindows.html)。   - 开始时间必须为22:00:00、02:00:00、06:00:00、10:00:00、14:00:00和18:00:00。   - 该参数不能单独为空，若该值为空，则结束时间也为空。系统分配一个默认开始时间02:00:00。
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// 维护时间窗结束时间，格式为HH:mm:ss。   - 维护时间窗开始和结束时间必须为指定的时间段，可参考查[询维护时间窗时间段](https://support.huaweicloud.com/api-kafka/ShowMaintainWindows.html)。   - 结束时间在开始时间基础上加四个小时，即当开始时间为22:00:00时，结束时间为02:00:00。   - 该参数不能单独为空，若该值为空，则开始时间也为空。系统分配一个默认结束时间06:00:00。
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// 安全组ID。
	SecurityGroupId *string `json:"security_group_id,omitempty"`
	// 容量阈值策略。 支持两种策略模式： - produce_reject: 生产受限 - time_base: 自动删除
	RetentionPolicy *UpdateInstanceReqRetentionPolicy `json:"retention_policy,omitempty"`
	// 企业项目。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o UpdateInstanceReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceReq struct{}"
	}

	return strings.Join([]string{"UpdateInstanceReq", string(data)}, " ")
}

type UpdateInstanceReqRetentionPolicy struct {
	value string
}

type UpdateInstanceReqRetentionPolicyEnum struct {
	PRODUCE_REJECT UpdateInstanceReqRetentionPolicy
	TIME_BASE      UpdateInstanceReqRetentionPolicy
}

func GetUpdateInstanceReqRetentionPolicyEnum() UpdateInstanceReqRetentionPolicyEnum {
	return UpdateInstanceReqRetentionPolicyEnum{
		PRODUCE_REJECT: UpdateInstanceReqRetentionPolicy{
			value: "produce_reject",
		},
		TIME_BASE: UpdateInstanceReqRetentionPolicy{
			value: "time_base",
		},
	}
}

func (c UpdateInstanceReqRetentionPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateInstanceReqRetentionPolicy) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
