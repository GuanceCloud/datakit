/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TopicAttribute struct {
	// 访问策略规范版本。目前只支持“2016-09-07”。
	Version string `json:"Version"`
	// 策略的唯一标识。不能为空。
	Id string `json:"Id"`
	// 访问策略是通过Statement语句来定义的。一个访问策略可包含一条或多条Statement语句。通过Statement语句向其他用户或云服务授权对主题的操作。
	Statement []Statement `json:"Statement"`
}

func (o TopicAttribute) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TopicAttribute struct{}"
	}

	return strings.Join([]string{"TopicAttribute", string(data)}, " ")
}
