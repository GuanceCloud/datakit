/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Container network parameters.
type ContainerNetwork struct {
	// 容器网络网段，建议使用网段10.0.0.0/12~19，172.16.0.0/16~19，192.168.0.0/16~19，如存在网段冲突，将自动重新选择。  当节点最大实例数为默认值110时，当前容器网段至少支持582个节点，此参数在集群创建后不可更改，请谨慎选择。
	Cidr *string `json:"cidr,omitempty"`
	// 容器网络类型（只可选择其一） - overlay_l2：通过OVS（OpenVSwitch）为容器构建的overlay _ l2网络。 - underlay_ipvlan：裸金属服务器使用ipvlan构建的Underlay的l2网络。 - vpc-router：使用ipvlan和自定义VPC路由为容器构建的Underlay的l2网络。 - eni：Yangtse网络，深度整合VPC原生ENI弹性网卡能力，采用VPC网段分配容器地址，支持ELB直通容器，享有高性能，创建CCE Turbo集群（公测中）时指定。  >   - 容器隧道网络（Overlay）：基于VXLAN技术实现的Overlay容器网络。VXLAN是将以太网报文封装成UDP报文进行隧道传输。容器网络是承载于VPC网络之上的Overlay网络平面，具有付出少量隧道封装性能损耗，获得了通用性强、互通性强、高级特性支持全面（例如Network Policy网络隔离）的优势，可以满足大多数应用需求。 >   - VPC网络：基于VPC网络的自定义路由，直接将容器网络承载于VPC网络之中。每个节点将会被分配固定大小的IP地址段。vpc-router网络由于没有隧道封装的消耗，容器网络性能相对于容器隧道网络有一定优势。vpc-router集群由于VPC路由中配置有容器网段与节点IP的路由，可以支持集群外直接访问容器实例等特殊场景。
	Mode ContainerNetworkMode `json:"mode"`
}

func (o ContainerNetwork) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ContainerNetwork struct{}"
	}

	return strings.Join([]string{"ContainerNetwork", string(data)}, " ")
}

type ContainerNetworkMode struct {
	value string
}

type ContainerNetworkModeEnum struct {
	OVERLAY_L2      ContainerNetworkMode
	VPC_ROUTER      ContainerNetworkMode
	UNDERLAY_IPVLAN ContainerNetworkMode
	ENI             ContainerNetworkMode
}

func GetContainerNetworkModeEnum() ContainerNetworkModeEnum {
	return ContainerNetworkModeEnum{
		OVERLAY_L2: ContainerNetworkMode{
			value: "overlay_l2",
		},
		VPC_ROUTER: ContainerNetworkMode{
			value: "vpc-router",
		},
		UNDERLAY_IPVLAN: ContainerNetworkMode{
			value: "underlay_ipvlan",
		},
		ENI: ContainerNetworkMode{
			value: "eni",
		},
	}
}

func (c ContainerNetworkMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ContainerNetworkMode) UnmarshalJSON(b []byte) error {
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
