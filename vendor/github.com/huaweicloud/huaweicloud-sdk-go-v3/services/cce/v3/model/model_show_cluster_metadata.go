/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ShowClusterMetadata struct {
	// 集群注解。此字段与创建时的annotations无必然关系，查询时根据查询参数返回集群相关信息存入该字段中。  当查询参数detail设置为true时，该注解包含集群下节点总数(totalNodesNumber)、正常节点数(activeNodesNumber)、CPU总量(totalNodesCPU)、内存总量(totalNodesMemory)和已安装插件名称(installedAddonInstances)。
	Annotations map[string]string `json:"annotations,omitempty"`
	// 集群创建时间，集群创建成功后自动生成，填写无效
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`
	// 集群标签，key/value对格式。  该字段值由系统自动生成，用于升级时前端识别集群支持的特性开关，用户指定无效，与创建时的labels无必然关系。
	Labels *string `json:"labels,omitempty"`
	// 集群名称。  命名规则：以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围4-128位，且不能以中划线(-)结尾。
	Name string `json:"name"`
	// 资源唯一标识，创建成功后自动生成，填写无效
	Uid *string `json:"uid,omitempty"`
	// 集群更新时间，集群创建成功后自动生成，填写无效
	UpdateTimestamp *string `json:"updateTimestamp,omitempty"`
}

func (o ShowClusterMetadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowClusterMetadata struct{}"
	}

	return strings.Join([]string{"ShowClusterMetadata", string(data)}, " ")
}
