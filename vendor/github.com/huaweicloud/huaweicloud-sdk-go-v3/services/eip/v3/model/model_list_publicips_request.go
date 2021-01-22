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
	Marker                    *string                                      `json:"marker,omitempty"`
	Offset                    *int32                                       `json:"offset,omitempty"`
	Limit                     *int32                                       `json:"limit,omitempty"`
	Fields                    *[]string                                    `json:"fields,omitempty"`
	SortKey                   *ListPublicipsRequestSortKey                 `json:"sort_key,omitempty"`
	SortDir                   *ListPublicipsRequestSortDir                 `json:"sort_dir,omitempty"`
	Id                        *[]string                                    `json:"id,omitempty"`
	IpVersion                 *[]ListPublicipsRequestIpVersion             `json:"ip_version,omitempty"`
	PublicIpAddress           *[]string                                    `json:"public_ip_address,omitempty"`
	PublicIpAddressLike       *string                                      `json:"public_ip_address_like,omitempty"`
	PublicIpv6Address         *[]string                                    `json:"public_ipv6_address,omitempty"`
	PublicIpv6AddressLike     *string                                      `json:"public_ipv6_address_like,omitempty"`
	Type                      *[]ListPublicipsRequestType                  `json:"type,omitempty"`
	NetworkType               *[]ListPublicipsRequestNetworkType           `json:"network_type,omitempty"`
	PublicipPoolName          *[]string                                    `json:"publicip_pool_name,omitempty"`
	Status                    *[]ListPublicipsRequestStatus                `json:"status,omitempty"`
	AliasLike                 *string                                      `json:"alias_like,omitempty"`
	Alias                     *[]string                                    `json:"alias,omitempty"`
	Description               *[]string                                    `json:"description,omitempty"`
	VnicPrivateIpAddress      *[]string                                    `json:"vnic.private_ip_address,omitempty"`
	VnicPrivateIpAddressLike  *string                                      `json:"vnic.private_ip_address_like,omitempty"`
	VnicDeviceId              *[]string                                    `json:"vnic.device_id,omitempty"`
	VnicDeviceOwner           *[]string                                    `json:"vnic.device_owner,omitempty"`
	VnicVpcId                 *[]string                                    `json:"vnic.vpc_id,omitempty"`
	VnicPortId                *[]string                                    `json:"vnic.port_id,omitempty"`
	VnicDeviceOwnerPrefixlike *string                                      `json:"vnic.device_owner_prefixlike,omitempty"`
	VnicInstanceType          *[]string                                    `json:"vnic.instance_type,omitempty"`
	VnicInstanceId            *[]string                                    `json:"vnic.instance_id,omitempty"`
	BandwidthId               *[]string                                    `json:"bandwidth.id,omitempty"`
	BandwidthName             *[]string                                    `json:"bandwidth.name,omitempty"`
	BandwidthNameLike         *[]string                                    `json:"bandwidth.name_like,omitempty"`
	BandwidthSize             *[]int32                                     `json:"bandwidth.size,omitempty"`
	BandwidthShareType        *[]ListPublicipsRequestBandwidthShareType    `json:"bandwidth.share_type,omitempty"`
	BandwidthChargeMode       *[]ListPublicipsRequestBandwidthChargeMode   `json:"bandwidth.charge_mode,omitempty"`
	BillingInfo               *[]string                                    `json:"billing_info,omitempty"`
	BillingMode               *ListPublicipsRequestBillingMode             `json:"billing_mode,omitempty"`
	AssociateInstanceType     *[]ListPublicipsRequestAssociateInstanceType `json:"associate_instance_type,omitempty"`
	AssociateInstanceId       *[]string                                    `json:"associate_instance_id,omitempty"`
	EnterpriseProjectId       *[]string                                    `json:"enterprise_project_id,omitempty"`
	GroupName                 *[]string                                    `json:"group_name,omitempty"`
}

func (o ListPublicipsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsRequest struct{}"
	}

	return strings.Join([]string{"ListPublicipsRequest", string(data)}, " ")
}

type ListPublicipsRequestSortKey struct {
	value string
}

type ListPublicipsRequestSortKeyEnum struct {
	ID                  ListPublicipsRequestSortKey
	PUBLIC_IP_ADDRESS   ListPublicipsRequestSortKey
	PUBLIC_IPV6_ADDRESS ListPublicipsRequestSortKey
	IP_VERSION          ListPublicipsRequestSortKey
	CREATED_AT          ListPublicipsRequestSortKey
	UPDATED_AT          ListPublicipsRequestSortKey
}

func GetListPublicipsRequestSortKeyEnum() ListPublicipsRequestSortKeyEnum {
	return ListPublicipsRequestSortKeyEnum{
		ID: ListPublicipsRequestSortKey{
			value: "id",
		},
		PUBLIC_IP_ADDRESS: ListPublicipsRequestSortKey{
			value: "public_ip_address",
		},
		PUBLIC_IPV6_ADDRESS: ListPublicipsRequestSortKey{
			value: "public_ipv6_address",
		},
		IP_VERSION: ListPublicipsRequestSortKey{
			value: "ip_version",
		},
		CREATED_AT: ListPublicipsRequestSortKey{
			value: "created_at",
		},
		UPDATED_AT: ListPublicipsRequestSortKey{
			value: "updated_at",
		},
	}
}

func (c ListPublicipsRequestSortKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestSortKey) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestSortDir struct {
	value string
}

type ListPublicipsRequestSortDirEnum struct {
	ASC  ListPublicipsRequestSortDir
	DESC ListPublicipsRequestSortDir
}

func GetListPublicipsRequestSortDirEnum() ListPublicipsRequestSortDirEnum {
	return ListPublicipsRequestSortDirEnum{
		ASC: ListPublicipsRequestSortDir{
			value: "asc",
		},
		DESC: ListPublicipsRequestSortDir{
			value: "desc",
		},
	}
}

func (c ListPublicipsRequestSortDir) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestSortDir) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestType struct {
	value string
}

type ListPublicipsRequestTypeEnum struct {
	EIP              ListPublicipsRequestType
	DUALSTACK        ListPublicipsRequestType
	DUALSTACK_SUBNET ListPublicipsRequestType
}

func GetListPublicipsRequestTypeEnum() ListPublicipsRequestTypeEnum {
	return ListPublicipsRequestTypeEnum{
		EIP: ListPublicipsRequestType{
			value: "EIP",
		},
		DUALSTACK: ListPublicipsRequestType{
			value: "DUALSTACK",
		},
		DUALSTACK_SUBNET: ListPublicipsRequestType{
			value: "DUALSTACK_SUBNET",
		},
	}
}

func (c ListPublicipsRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestType) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestNetworkType struct {
	value string
}

type ListPublicipsRequestNetworkTypeEnum struct {
	E_5_TELCOM  ListPublicipsRequestNetworkType
	E_5_UNION   ListPublicipsRequestNetworkType
	E_5_BGP     ListPublicipsRequestNetworkType
	E_5_SBGP    ListPublicipsRequestNetworkType
	E_5_IPV6    ListPublicipsRequestNetworkType
	E_5_GRAYBGP ListPublicipsRequestNetworkType
}

func GetListPublicipsRequestNetworkTypeEnum() ListPublicipsRequestNetworkTypeEnum {
	return ListPublicipsRequestNetworkTypeEnum{
		E_5_TELCOM: ListPublicipsRequestNetworkType{
			value: "5_telcom",
		},
		E_5_UNION: ListPublicipsRequestNetworkType{
			value: "5_union",
		},
		E_5_BGP: ListPublicipsRequestNetworkType{
			value: "5_bgp",
		},
		E_5_SBGP: ListPublicipsRequestNetworkType{
			value: "5_sbgp",
		},
		E_5_IPV6: ListPublicipsRequestNetworkType{
			value: "5_ipv6",
		},
		E_5_GRAYBGP: ListPublicipsRequestNetworkType{
			value: "5_graybgp",
		},
	}
}

func (c ListPublicipsRequestNetworkType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestNetworkType) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestStatus struct {
	value string
}

type ListPublicipsRequestStatusEnum struct {
	FREEZED ListPublicipsRequestStatus
	DOWN    ListPublicipsRequestStatus
	ACTIVE  ListPublicipsRequestStatus
	ERROR   ListPublicipsRequestStatus
}

func GetListPublicipsRequestStatusEnum() ListPublicipsRequestStatusEnum {
	return ListPublicipsRequestStatusEnum{
		FREEZED: ListPublicipsRequestStatus{
			value: "FREEZED",
		},
		DOWN: ListPublicipsRequestStatus{
			value: "DOWN",
		},
		ACTIVE: ListPublicipsRequestStatus{
			value: "ACTIVE",
		},
		ERROR: ListPublicipsRequestStatus{
			value: "ERROR",
		},
	}
}

func (c ListPublicipsRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestStatus) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestBandwidthShareType struct {
	value string
}

type ListPublicipsRequestBandwidthShareTypeEnum struct {
	PER   ListPublicipsRequestBandwidthShareType
	WHOLE ListPublicipsRequestBandwidthShareType
}

func GetListPublicipsRequestBandwidthShareTypeEnum() ListPublicipsRequestBandwidthShareTypeEnum {
	return ListPublicipsRequestBandwidthShareTypeEnum{
		PER: ListPublicipsRequestBandwidthShareType{
			value: "PER",
		},
		WHOLE: ListPublicipsRequestBandwidthShareType{
			value: "WHOLE",
		},
	}
}

func (c ListPublicipsRequestBandwidthShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestBandwidthShareType) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestBandwidthChargeMode struct {
	value string
}

type ListPublicipsRequestBandwidthChargeModeEnum struct {
	BANDWIDTH     ListPublicipsRequestBandwidthChargeMode
	TRAFFIC       ListPublicipsRequestBandwidthChargeMode
	E_95PEAK_PLUS ListPublicipsRequestBandwidthChargeMode
}

func GetListPublicipsRequestBandwidthChargeModeEnum() ListPublicipsRequestBandwidthChargeModeEnum {
	return ListPublicipsRequestBandwidthChargeModeEnum{
		BANDWIDTH: ListPublicipsRequestBandwidthChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: ListPublicipsRequestBandwidthChargeMode{
			value: "traffic",
		},
		E_95PEAK_PLUS: ListPublicipsRequestBandwidthChargeMode{
			value: "95peak_plus",
		},
	}
}

func (c ListPublicipsRequestBandwidthChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestBandwidthChargeMode) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestBillingMode struct {
	value string
}

type ListPublicipsRequestBillingModeEnum struct {
	YEARLY_MONTHLY ListPublicipsRequestBillingMode
	PAY_PER_USE    ListPublicipsRequestBillingMode
}

func GetListPublicipsRequestBillingModeEnum() ListPublicipsRequestBillingModeEnum {
	return ListPublicipsRequestBillingModeEnum{
		YEARLY_MONTHLY: ListPublicipsRequestBillingMode{
			value: "YEARLY_MONTHLY",
		},
		PAY_PER_USE: ListPublicipsRequestBillingMode{
			value: "PAY_PER_USE",
		},
	}
}

func (c ListPublicipsRequestBillingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestBillingMode) UnmarshalJSON(b []byte) error {
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

type ListPublicipsRequestAssociateInstanceType struct {
	value string
}

type ListPublicipsRequestAssociateInstanceTypeEnum struct {
	PORT  ListPublicipsRequestAssociateInstanceType
	NATGW ListPublicipsRequestAssociateInstanceType
	ELB   ListPublicipsRequestAssociateInstanceType
	VPN   ListPublicipsRequestAssociateInstanceType
	ELBV1 ListPublicipsRequestAssociateInstanceType
}

func GetListPublicipsRequestAssociateInstanceTypeEnum() ListPublicipsRequestAssociateInstanceTypeEnum {
	return ListPublicipsRequestAssociateInstanceTypeEnum{
		PORT: ListPublicipsRequestAssociateInstanceType{
			value: "PORT",
		},
		NATGW: ListPublicipsRequestAssociateInstanceType{
			value: "NATGW",
		},
		ELB: ListPublicipsRequestAssociateInstanceType{
			value: "ELB",
		},
		VPN: ListPublicipsRequestAssociateInstanceType{
			value: "VPN",
		},
		ELBV1: ListPublicipsRequestAssociateInstanceType{
			value: "ELBV1",
		},
	}
}

func (c ListPublicipsRequestAssociateInstanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsRequestAssociateInstanceType) UnmarshalJSON(b []byte) error {
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
