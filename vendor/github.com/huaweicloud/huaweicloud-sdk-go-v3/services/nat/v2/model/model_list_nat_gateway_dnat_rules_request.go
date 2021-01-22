/*
 * NAT
 *
 * Open Api of Public Nat.
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// Request Object
type ListNatGatewayDnatRulesRequest struct {
	AdminStateUp        *bool                                   `json:"admin_state_up,omitempty"`
	ExternalServicePort *int32                                  `json:"external_service_port,omitempty"`
	FloatingIpAddress   *string                                 `json:"floating_ip_address,omitempty"`
	Status              *[]ListNatGatewayDnatRulesRequestStatus `json:"status,omitempty"`
	FloatingIpId        *string                                 `json:"floating_ip_id,omitempty"`
	InternalServicePort *int32                                  `json:"internal_service_port,omitempty"`
	Limit               *int32                                  `json:"limit,omitempty"`
	Id                  *string                                 `json:"id,omitempty"`
	Description         *string                                 `json:"description,omitempty"`
	CreatedAt           *sdktime.SdkTime                        `json:"created_at,omitempty"`
	NatGatewayId        *[]string                               `json:"nat_gateway_id,omitempty"`
	PortId              *string                                 `json:"port_id,omitempty"`
	PrivateIp           *string                                 `json:"private_ip,omitempty"`
	Protocol            *[]string                               `json:"protocol,omitempty"`
}

func (o ListNatGatewayDnatRulesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNatGatewayDnatRulesRequest struct{}"
	}

	return strings.Join([]string{"ListNatGatewayDnatRulesRequest", string(data)}, " ")
}

type ListNatGatewayDnatRulesRequestStatus struct {
	value string
}

type ListNatGatewayDnatRulesRequestStatusEnum struct {
	ACTIVE         ListNatGatewayDnatRulesRequestStatus
	PENDING_CREATE ListNatGatewayDnatRulesRequestStatus
	PENDING_UPDATE ListNatGatewayDnatRulesRequestStatus
	PENDING_DELETE ListNatGatewayDnatRulesRequestStatus
	EIP_FREEZED    ListNatGatewayDnatRulesRequestStatus
	INACTIVE       ListNatGatewayDnatRulesRequestStatus
}

func GetListNatGatewayDnatRulesRequestStatusEnum() ListNatGatewayDnatRulesRequestStatusEnum {
	return ListNatGatewayDnatRulesRequestStatusEnum{
		ACTIVE: ListNatGatewayDnatRulesRequestStatus{
			value: "ACTIVE",
		},
		PENDING_CREATE: ListNatGatewayDnatRulesRequestStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: ListNatGatewayDnatRulesRequestStatus{
			value: "PENDING_UPDATE",
		},
		PENDING_DELETE: ListNatGatewayDnatRulesRequestStatus{
			value: "PENDING_DELETE",
		},
		EIP_FREEZED: ListNatGatewayDnatRulesRequestStatus{
			value: "EIP_FREEZED",
		},
		INACTIVE: ListNatGatewayDnatRulesRequestStatus{
			value: "INACTIVE",
		},
	}
}

func (c ListNatGatewayDnatRulesRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListNatGatewayDnatRulesRequestStatus) UnmarshalJSON(b []byte) error {
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
