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

type NodePoolStatus struct {
	// 当前节点池中节点数量
	CurrentNode *int32 `json:"currentNode,omitempty"`
	// 节点池删除时的 JobID
	JobId *string `json:"jobId,omitempty"`
	// 节点池状态，可为空。
	Phase *NodePoolStatusPhase `json:"phase,omitempty"`
}

func (o NodePoolStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NodePoolStatus struct{}"
	}

	return strings.Join([]string{"NodePoolStatus", string(data)}, " ")
}

type NodePoolStatusPhase struct {
	value string
}

type NodePoolStatusPhaseEnum struct {
	SYNCHRONIZING NodePoolStatusPhase
	SYNCHRONIZED  NodePoolStatusPhase
	SOLD_OUT      NodePoolStatusPhase
	DELETING      NodePoolStatusPhase
	ERROR         NodePoolStatusPhase
}

func GetNodePoolStatusPhaseEnum() NodePoolStatusPhaseEnum {
	return NodePoolStatusPhaseEnum{
		SYNCHRONIZING: NodePoolStatusPhase{
			value: "Synchronizing",
		},
		SYNCHRONIZED: NodePoolStatusPhase{
			value: "Synchronized",
		},
		SOLD_OUT: NodePoolStatusPhase{
			value: "SoldOut",
		},
		DELETING: NodePoolStatusPhase{
			value: "Deleting",
		},
		ERROR: NodePoolStatusPhase{
			value: "Error",
		},
	}
}

func (c NodePoolStatusPhase) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NodePoolStatusPhase) UnmarshalJSON(b []byte) error {
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
