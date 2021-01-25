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

// 编排flow详情，描述流水线内各阶段任务的串并行关系。map类型数据，key为阶段名字，默认第一阶段initial，最后阶段为final，其余名字以'state_数字'标识。value为该阶段内任务(以'Task_数字'标识)以及后续阶段的标识。本字段为描述流水线基础编排数据之一，建议可通过流水线真实界面基于模板创建接口中获取
type FlowItem struct {
}

func (o FlowItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FlowItem struct{}"
	}

	return strings.Join([]string{"FlowItem", string(data)}, " ")
}
