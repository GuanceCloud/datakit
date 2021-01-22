/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type Rule struct {
	// template是date_rule时，设置params[\"days\"] template是tag_rule时，设置params[\"num\"]
	Params *interface{} `json:"params"`
	// 例外镜像
	TagSelectors []TagSelector `json:"tag_selectors"`
	// 回收类型，date_rule、tag_rule
	Template RuleTemplate `json:"template"`
}

func (o Rule) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Rule struct{}"
	}

	return strings.Join([]string{"Rule", string(data)}, " ")
}

type RuleTemplate struct {
	value string
}

type RuleTemplateEnum struct {
	DATE_RULE RuleTemplate
	TAG_RULE  RuleTemplate
}

func GetRuleTemplateEnum() RuleTemplateEnum {
	return RuleTemplateEnum{
		DATE_RULE: RuleTemplate{
			value: "date_rule",
		},
		TAG_RULE: RuleTemplate{
			value: "tag_rule",
		},
	}
}

func (c RuleTemplate) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RuleTemplate) UnmarshalJSON(b []byte) error {
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
