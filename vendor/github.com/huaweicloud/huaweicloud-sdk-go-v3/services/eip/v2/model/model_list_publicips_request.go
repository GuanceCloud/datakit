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

// Request Object
type ListPublicipsRequest struct {
	Marker              *string                        `json:"marker,omitempty"`
	Limit               *int32                         `json:"limit,omitempty"`
	IpVersion           *ListPublicipsRequestIpVersion `json:"ip_version,omitempty"`
	EnterpriseProjectId *string                        `json:"enterprise_project_id,omitempty"`
	PortId              *[]string                      `json:"port_id,omitempty"`
	PublicIpAddress     *[]string                      `json:"public_ip_address,omitempty"`
	PrivateIpAddress    *[]string                      `json:"private_ip_address,omitempty"`
	Id                  *[]string                      `json:"id,omitempty"`
}

func (o ListPublicipsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsRequest struct{}"
	}

	return strings.Join([]string{"ListPublicipsRequest", string(data)}, " ")
}

type ListPublicipsRequestIpVersion struct {
	value int32
}

type ListPublicipsRequestIpVersionEnum struct {
	E_4 ListPublicipsRequestIpVersion
	E_6 ListPublicipsRequestIpVersion
}

func GetListPublicipsRequestIpVersionEnum() ListPublicipsRequestIpVersionEnum {
	return ListPublicipsRequestIpVersionEnum{
		E_4: ListPublicipsRequestIpVersion{
			value: 4,
		}, E_6: ListPublicipsRequestIpVersion{
			value: 6,
		},
	}
}

func (c ListPublicipsRequestIpVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestIpVersion) UnmarshalJSON(b []byte) error {
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
