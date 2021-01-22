/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// PublicipInfo对象
type PublicipInfoResp struct {
	// 功能说明：弹性公网IP或者IPv6端口的地址
	PublicipAddress *string `json:"publicip_address,omitempty"`
	// 功能说明：带宽对应的弹性公网IP或者IPv6端口的唯一标识
	PublicipId *string `json:"publicip_id,omitempty"`
	// 功能说明：弹性公网IP或者IPv6端口的类型  取值范围：5_telcom（电信），5_union（联通），5_bgp（全动态BGP），5_sbgp（静态BGP），5_ipv6  东北-大连：5_telcom、5_union  华南-广州：5_bgp、5_sbgp  华东-上海二：5_bgp、5_sbgp  华北-北京一：5_bgp、5_sbgp、5_ipv6  亚太-香港：5_bgp  亚太-曼谷：5_bgp  亚太-新加坡：5_bgp  非洲-约翰内斯堡：5_bgp  西南-贵阳一：5_bgp、5_sbgp  华北-北京四：5_bgp、5_sbgp  约束：必须是系统具体支持的类型
	PublicipType *string `json:"publicip_type,omitempty"`
	// 功能说明：IPv4时无此字段，IPv6时为申请到的弹性公网IP地址
	Publicipv6Address *string `json:"publicipv6_address,omitempty"`
	// IP版本信息  取值范围：  4：IPv4  6：IPv6
	IpVersion *PublicipInfoRespIpVersion `json:"ip_version,omitempty"`
}

func (o PublicipInfoResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublicipInfoResp struct{}"
	}

	return strings.Join([]string{"PublicipInfoResp", string(data)}, " ")
}

type PublicipInfoRespIpVersion struct {
	value int32
}

type PublicipInfoRespIpVersionEnum struct {
	E_4 PublicipInfoRespIpVersion
	E_6 PublicipInfoRespIpVersion
}

func GetPublicipInfoRespIpVersionEnum() PublicipInfoRespIpVersionEnum {
	return PublicipInfoRespIpVersionEnum{
		E_4: PublicipInfoRespIpVersion{
			value: 4,
		}, E_6: PublicipInfoRespIpVersion{
			value: 6,
		},
	}
}

func (c PublicipInfoRespIpVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipInfoRespIpVersion) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}
