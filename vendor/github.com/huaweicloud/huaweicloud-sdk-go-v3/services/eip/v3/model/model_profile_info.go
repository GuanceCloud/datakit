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

// 公网IP元数据, EIP服务内部使用，不对外开放
type ProfileInfo struct {
	// 公网IP附属的5_xxx网络（如5_bgp）中的port_id
	LocalNetworkPort *string `json:"local_network_port,omitempty"`
	// 标识公网IP是否是和虚机一起创建的。true-独立创建；false-和虚机一起创建
	Standalone *bool `json:"standalone,omitempty"`
	// 云服务标识公网IP创建进度, EIP服务内部使用。
	NotifyStatus *ProfileInfoNotifyStatus `json:"notify_status,omitempty"`
	// 公网IP创建时间
	CreateTime *string `json:"create_time,omitempty"`
	// 该字段仅仅用于表示eip的bgp类型是否是真实的静态sbgp * 1. 如果为true，则该eip可以切换bgp类型 * 2. 如果为false，则该eip不可以切换bgp类型
	FakeNetworkType *bool `json:"fake_network_type,omitempty"`
	// 标识IP是和哪类资源一起购买的
	CreateSource *ProfileInfoCreateSource `json:"create_source,omitempty"`
	// 标识和公网IP一起购买的ecs的id
	EcsId *string `json:"ecs_id,omitempty"`
	// 公网IP加锁状态, eg:\"POLICE,LOCKED\"。POLICE-公安冻结；LOCKED-普通冻结；普通冻结细分状态：ARREAR-欠费；DELABLE-可删除；
	LockStatus *string `json:"lock_status,omitempty"`
	// 公网IP冻结状态。
	FreezedStatus *ProfileInfoFreezedStatus `json:"freezed_status,omitempty"`
	BandwithInfo  *BandwidthInfoResp        `json:"bandwith_info,omitempty"`
}

func (o ProfileInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ProfileInfo struct{}"
	}

	return strings.Join([]string{"ProfileInfo", string(data)}, " ")
}

type ProfileInfoNotifyStatus struct {
	value string
}

type ProfileInfoNotifyStatusEnum struct {
	PENDING_CREATE ProfileInfoNotifyStatus
	PENDING_UPDATE ProfileInfoNotifyStatus
	NOTIFYING      ProfileInfoNotifyStatus
	NOTIFYED       ProfileInfoNotifyStatus
	NOTIFY_DELETE  ProfileInfoNotifyStatus
}

func GetProfileInfoNotifyStatusEnum() ProfileInfoNotifyStatusEnum {
	return ProfileInfoNotifyStatusEnum{
		PENDING_CREATE: ProfileInfoNotifyStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: ProfileInfoNotifyStatus{
			value: "PENDING_UPDATE",
		},
		NOTIFYING: ProfileInfoNotifyStatus{
			value: "NOTIFYING",
		},
		NOTIFYED: ProfileInfoNotifyStatus{
			value: "NOTIFYED",
		},
		NOTIFY_DELETE: ProfileInfoNotifyStatus{
			value: "NOTIFY_DELETE",
		},
	}
}

func (c ProfileInfoNotifyStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ProfileInfoNotifyStatus) UnmarshalJSON(b []byte) error {
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

type ProfileInfoCreateSource struct {
	value string
}

type ProfileInfoCreateSourceEnum struct {
	ECS ProfileInfoCreateSource
}

func GetProfileInfoCreateSourceEnum() ProfileInfoCreateSourceEnum {
	return ProfileInfoCreateSourceEnum{
		ECS: ProfileInfoCreateSource{
			value: "ecs",
		},
	}
}

func (c ProfileInfoCreateSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ProfileInfoCreateSource) UnmarshalJSON(b []byte) error {
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

type ProfileInfoFreezedStatus struct {
	value string
}

type ProfileInfoFreezedStatusEnum struct {
	FREEZED   ProfileInfoFreezedStatus
	UNFREEZED ProfileInfoFreezedStatus
}

func GetProfileInfoFreezedStatusEnum() ProfileInfoFreezedStatusEnum {
	return ProfileInfoFreezedStatusEnum{
		FREEZED: ProfileInfoFreezedStatus{
			value: "FREEZED",
		},
		UNFREEZED: ProfileInfoFreezedStatus{
			value: "UNFREEZED",
		},
	}
}

func (c ProfileInfoFreezedStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ProfileInfoFreezedStatus) UnmarshalJSON(b []byte) error {
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
