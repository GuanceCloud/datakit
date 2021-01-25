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
type ListNatGatewaySnatRulesRequest struct {
	AdminStateUp      *bool                                 `json:"admin_state_up,omitempty"`
	Cidr              *string                               `json:"cidr,omitempty"`
	Limit             *int32                                `json:"limit,omitempty"`
	FloatingIpAddress *string                               `json:"floating_ip_address,omitempty"`
	FloatingIpId      *string                               `json:"floating_ip_id,omitempty"`
	Id                *string                               `json:"id,omitempty"`
	Description       *string                               `json:"description,omitempty"`
	CreatedAt         *sdktime.SdkTime                      `json:"created_at,omitempty"`
	NatGatewayId      *[]string                             `json:"nat_gateway_id,omitempty"`
	NetworkId         *string                               `json:"network_id,omitempty"`
	SourceType        *int32                                `json:"source_type,omitempty"`
	Status            *ListNatGatewaySnatRulesRequestStatus `json:"status,omitempty"`
}

func (o ListNatGatewaySnatRulesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNatGatewaySnatRulesRequest struct{}"
	}

	return strings.Join([]string{"ListNatGatewaySnatRulesRequest", string(data)}, " ")
}

type ListNatGatewaySnatRulesRequestStatus struct {
	value string
}

type ListNatGatewaySnatRulesRequestStatusEnum struct {
	ACTIVE         ListNatGatewaySnatRulesRequestStatus
	PENDING_CREATE ListNatGatewaySnatRulesRequestStatus
	PENDING_UPDATE ListNatGatewaySnatRulesRequestStatus
	PENDING_DELETE ListNatGatewaySnatRulesRequestStatus
	EIP_FREEZED    ListNatGatewaySnatRulesRequestStatus
	INACTIVE       ListNatGatewaySnatRulesRequestStatus
}

func GetListNatGatewaySnatRulesRequestStatusEnum() ListNatGatewaySnatRulesRequestStatusEnum {
	return ListNatGatewaySnatRulesRequestStatusEnum{
		ACTIVE: ListNatGatewaySnatRulesRequestStatus{
			value: "ACTIVE",
		},
		PENDING_CREATE: ListNatGatewaySnatRulesRequestStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: ListNatGatewaySnatRulesRequestStatus{
			value: "PENDING_UPDATE",
		},
		PENDING_DELETE: ListNatGatewaySnatRulesRequestStatus{
			value: "PENDING_DELETE",
		},
		EIP_FREEZED: ListNatGatewaySnatRulesRequestStatus{
			value: "EIP_FREEZED",
		},
		INACTIVE: ListNatGatewaySnatRulesRequestStatus{
			value: "INACTIVE",
		},
	}
}

func (c ListNatGatewaySnatRulesRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListNatGatewaySnatRulesRequestStatus) UnmarshalJSON(b []byte) error {
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
