/*
 * DIS
 *
 * DIS v1 API
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
type DeleteCheckpointRequest struct {
	StreamName     string                                `json:"stream_name"`
	AppName        string                                `json:"app_name"`
	CheckpointType DeleteCheckpointRequestCheckpointType `json:"checkpoint_type"`
	PartitionId    *string                               `json:"partition_id,omitempty"`
}

func (o DeleteCheckpointRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteCheckpointRequest struct{}"
	}

	return strings.Join([]string{"DeleteCheckpointRequest", string(data)}, " ")
}

type DeleteCheckpointRequestCheckpointType struct {
	value string
}

type DeleteCheckpointRequestCheckpointTypeEnum struct {
	LAST_READ DeleteCheckpointRequestCheckpointType
}

func GetDeleteCheckpointRequestCheckpointTypeEnum() DeleteCheckpointRequestCheckpointTypeEnum {
	return DeleteCheckpointRequestCheckpointTypeEnum{
		LAST_READ: DeleteCheckpointRequestCheckpointType{
			value: "LAST_READ",
		},
	}
}

func (c DeleteCheckpointRequestCheckpointType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteCheckpointRequestCheckpointType) UnmarshalJSON(b []byte) error {
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
