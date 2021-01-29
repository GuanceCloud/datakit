/*
 * VPC
 *
 * VPC Open API
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
type ListPortsRequest struct {
	Name                *string                      `json:"name,omitempty"`
	Id                  *string                      `json:"id,omitempty"`
	Limit               *int32                       `json:"limit,omitempty"`
	AdminStateUp        *bool                        `json:"admin_state_up,omitempty"`
	NetworkId           *string                      `json:"network_id,omitempty"`
	MacAddress          *string                      `json:"mac_address,omitempty"`
	DeviceId            *string                      `json:"device_id,omitempty"`
	DeviceOwner         *ListPortsRequestDeviceOwner `json:"device_owner,omitempty"`
	Status              *ListPortsRequestStatus      `json:"status,omitempty"`
	Marker              *string                      `json:"marker,omitempty"`
	FixedIps            *string                      `json:"fixed_ips,omitempty"`
	EnterpriseProjectId *string                      `json:"enterprise_project_id,omitempty"`
	ZoneId              *string                      `json:"zone_id,omitempty"`
}

func (o ListPortsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPortsRequest struct{}"
	}

	return strings.Join([]string{"ListPortsRequest", string(data)}, " ")
}

type ListPortsRequestDeviceOwner struct {
	value string
}

type ListPortsRequestDeviceOwnerEnum struct {
	NETWORKDHCP                         ListPortsRequestDeviceOwner
	NETWORKVIP_PORT                     ListPortsRequestDeviceOwner
	NETWORKROUTER_INTERFACE_DISTRIBUTED ListPortsRequestDeviceOwner
	NETWORKROUTER_CENTRALIZED_SNAT      ListPortsRequestDeviceOwner
}

func GetListPortsRequestDeviceOwnerEnum() ListPortsRequestDeviceOwnerEnum {
	return ListPortsRequestDeviceOwnerEnum{
		NETWORKDHCP: ListPortsRequestDeviceOwner{
			value: "network:dhcp",
		},
		NETWORKVIP_PORT: ListPortsRequestDeviceOwner{
			value: "network:VIP_PORT",
		},
		NETWORKROUTER_INTERFACE_DISTRIBUTED: ListPortsRequestDeviceOwner{
			value: "network:router_interface_distributed",
		},
		NETWORKROUTER_CENTRALIZED_SNAT: ListPortsRequestDeviceOwner{
			value: "network:router_centralized_snat",
		},
	}
}

func (c ListPortsRequestDeviceOwner) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPortsRequestDeviceOwner) UnmarshalJSON(b []byte) error {
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

type ListPortsRequestStatus struct {
	value string
}

type ListPortsRequestStatusEnum struct {
	ACTIVE ListPortsRequestStatus
	BUILD  ListPortsRequestStatus
	DOWN   ListPortsRequestStatus
}

func GetListPortsRequestStatusEnum() ListPortsRequestStatusEnum {
	return ListPortsRequestStatusEnum{
		ACTIVE: ListPortsRequestStatus{
			value: "ACTIVE",
		},
		BUILD: ListPortsRequestStatus{
			value: "BUILD",
		},
		DOWN: ListPortsRequestStatus{
			value: "DOWN",
		},
	}
}

func (c ListPortsRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPortsRequestStatus) UnmarshalJSON(b []byte) error {
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
