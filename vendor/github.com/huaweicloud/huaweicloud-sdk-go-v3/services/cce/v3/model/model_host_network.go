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

// Node network parameters.
type HostNetwork struct {
	// 节点安全组ID，创建时指定无效
	SecurityGroup *string `json:"SecurityGroup,omitempty"`
	// 用于创建裸金属节点的高速网络的子网ID。该值在[创建高速网络（可选）](https://support.huaweicloud.com/api-cce/cce_02_0270.html)中获取。创建裸金属集群时该参数为必选参数。
	HighwaySubnet *string `json:"highwaySubnet,omitempty"`
	// 用于创建控制节点的subnet的网络ID。获取方法如下：   - 方法1：登录虚拟私有云服务的控制台界面，单击VPC下的子网，进入子网详情页面，查找网络ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询子网列表](https://support.huaweicloud.com/api-vpc/vpc_subnet01_0003.html)
	Subnet string `json:"subnet"`
	// 用于创建控制节点的VPC的ID。该值在[创建VPC和子网](https://support.huaweicloud.com/api-cce/cce_02_0100.html)中获取。  获取方法如下：   - 方法1：登录虚拟私有云服务的控制台界面，在虚拟私有云的详情页面查找VPC ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询VPC列表](https://support.huaweicloud.com/api-vpc/vpc_api01_0003.html)   >当前vpc-router容器网络模型不支持对接含拓展网段的VPC。 >若您的用户类型为企业用户，则需要保证vpc所属的企业项目ID和集群创建时选择的企业项目ID一致。集群所属的企业项目ID通过extendParam字段下的enterpriseProjectId体现，该值默认为\"0\"，表示默认的企业项目。
	Vpc string `json:"vpc"`
}

func (o HostNetwork) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "HostNetwork struct{}"
	}

	return strings.Join([]string{"HostNetwork", string(data)}, " ")
}
