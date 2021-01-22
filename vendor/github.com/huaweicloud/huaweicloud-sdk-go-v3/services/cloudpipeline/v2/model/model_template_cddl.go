/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 创建流水线接口入参
type TemplateCddl struct {
	Flow *FlowItem `json:"flow"`
	// 子任务states，map类型数据
	States   map[string]TemplateState `json:"states"`
	Workflow *Workflow                `json:"workflow"`
}

func (o TemplateCddl) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateCddl struct{}"
	}

	return strings.Join([]string{"TemplateCddl", string(data)}, " ")
}
