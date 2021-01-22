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
type ListNatGatewaysRequest struct {
	TenantId            *string                         `json:"tenant_id,omitempty"`
	Id                  *string                         `json:"id,omitempty"`
	EnterpriseProjectId *string                         `json:"enterprise_project_id,omitempty"`
	Description         *string                         `json:"description,omitempty"`
	CreatedAt           *sdktime.SdkTime                `json:"created_at,omitempty"`
	Name                *string                         `json:"name,omitempty"`
	Status              *[]ListNatGatewaysRequestStatus `json:"status,omitempty"`
	Spec                *[]ListNatGatewaysRequestSpec   `json:"spec,omitempty"`
	AdminStateUp        *bool                           `json:"admin_state_up,omitempty"`
	InternalNetworkId   *string                         `json:"internal_network_id,omitempty"`
	RouterId            *string                         `json:"router_id,omitempty"`
	Limit               *int32                          `json:"limit,omitempty"`
}

func (o ListNatGatewaysRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNatGatewaysRequest struct{}"
	}

	return strings.Join([]string{"ListNatGatewaysRequest", string(data)}, " ")
}

type ListNatGatewaysRequestStatus struct {
	value string
}

type ListNatGatewaysRequestStatusEnum struct {
	ACTIVE         ListNatGatewaysRequestStatus
	PENDING_CREATE ListNatGatewaysRequestStatus
	PENDING_UPDATE ListNatGatewaysRequestStatus
	PENDING_DELETE ListNatGatewaysRequestStatus
	INACTIVE       ListNatGatewaysRequestStatus
}

func GetListNatGatewaysRequestStatusEnum() ListNatGatewaysRequestStatusEnum {
	return ListNatGatewaysRequestStatusEnum{
		ACTIVE: ListNatGatewaysRequestStatus{
			value: "ACTIVE",
		},
		PENDING_CREATE: ListNatGatewaysRequestStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: ListNatGatewaysRequestStatus{
			value: "PENDING_UPDATE",
		},
		PENDING_DELETE: ListNatGatewaysRequestStatus{
			value: "PENDING_DELETE",
		},
		INACTIVE: ListNatGatewaysRequestStatus{
			value: "INACTIVE",
		},
	}
}

func (c ListNatGatewaysRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListNatGatewaysRequestStatus) UnmarshalJSON(b []byte) error {
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

type ListNatGatewaysRequestSpec struct {
	value string
}

type ListNatGatewaysRequestSpecEnum struct {
	E_1 ListNatGatewaysRequestSpec
	E_2 ListNatGatewaysRequestSpec
	E_3 ListNatGatewaysRequestSpec
	E_4 ListNatGatewaysRequestSpec
}

func GetListNatGatewaysRequestSpecEnum() ListNatGatewaysRequestSpecEnum {
	return ListNatGatewaysRequestSpecEnum{
		E_1: ListNatGatewaysRequestSpec{
			value: "1",
		},
		E_2: ListNatGatewaysRequestSpec{
			value: "2",
		},
		E_3: ListNatGatewaysRequestSpec{
			value: "3",
		},
		E_4: ListNatGatewaysRequestSpec{
			value: "4",
		},
	}
}

func (c ListNatGatewaysRequestSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListNatGatewaysRequestSpec) UnmarshalJSON(b []byte) error {
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
