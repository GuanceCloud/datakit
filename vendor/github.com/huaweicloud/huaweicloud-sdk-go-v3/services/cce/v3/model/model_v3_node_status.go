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

type V3NodeStatus struct {
	DeleteStatus *DeleteStatus `json:"deleteStatus,omitempty"`
	// 创建或删除时的任务ID。
	JobID *string `json:"jobID,omitempty"`
	// 节点状态。
	Phase *V3NodeStatusPhase `json:"phase,omitempty"`
	// 节点主网卡私有网段IP地址。
	PrivateIP *string `json:"privateIP,omitempty"`
	// 节点弹性公网IP地址。
	PublicIP *string `json:"publicIP,omitempty"`
	// 底层云服务器或裸金属节点ID。
	ServerId *string `json:"serverId,omitempty"`
}

func (o V3NodeStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V3NodeStatus struct{}"
	}

	return strings.Join([]string{"V3NodeStatus", string(data)}, " ")
}

type V3NodeStatusPhase struct {
	value string
}

type V3NodeStatusPhaseEnum struct {
	BUILD      V3NodeStatusPhase
	INSTALLING V3NodeStatusPhase
	INSTALLED  V3NodeStatusPhase
	SHUT_DOWN  V3NodeStatusPhase
	UPGRADING  V3NodeStatusPhase
	ACTIVE     V3NodeStatusPhase
	ABNORMAL   V3NodeStatusPhase
	DELETING   V3NodeStatusPhase
	ERROR      V3NodeStatusPhase
}

func GetV3NodeStatusPhaseEnum() V3NodeStatusPhaseEnum {
	return V3NodeStatusPhaseEnum{
		BUILD: V3NodeStatusPhase{
			value: "Build",
		},
		INSTALLING: V3NodeStatusPhase{
			value: "Installing",
		},
		INSTALLED: V3NodeStatusPhase{
			value: "Installed",
		},
		SHUT_DOWN: V3NodeStatusPhase{
			value: "ShutDown",
		},
		UPGRADING: V3NodeStatusPhase{
			value: "Upgrading",
		},
		ACTIVE: V3NodeStatusPhase{
			value: "Active",
		},
		ABNORMAL: V3NodeStatusPhase{
			value: "Abnormal",
		},
		DELETING: V3NodeStatusPhase{
			value: "Deleting",
		},
		ERROR: V3NodeStatusPhase{
			value: "Error",
		},
	}
}

func (c V3NodeStatusPhase) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *V3NodeStatusPhase) UnmarshalJSON(b []byte) error {
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
