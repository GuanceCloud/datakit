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
type GetCheckpointRequest struct {
	StreamName     string                             `json:"stream_name"`
	PartitionId    string                             `json:"partition_id"`
	AppName        string                             `json:"app_name"`
	CheckpointType GetCheckpointRequestCheckpointType `json:"checkpoint_type"`
}

func (o GetCheckpointRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetCheckpointRequest struct{}"
	}

	return strings.Join([]string{"GetCheckpointRequest", string(data)}, " ")
}

type GetCheckpointRequestCheckpointType struct {
	value string
}

type GetCheckpointRequestCheckpointTypeEnum struct {
	LAST_READ GetCheckpointRequestCheckpointType
}

func GetGetCheckpointRequestCheckpointTypeEnum() GetCheckpointRequestCheckpointTypeEnum {
	return GetCheckpointRequestCheckpointTypeEnum{
		LAST_READ: GetCheckpointRequestCheckpointType{
			value: "LAST_READ",
		},
	}
}

func (c GetCheckpointRequestCheckpointType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *GetCheckpointRequestCheckpointType) UnmarshalJSON(b []byte) error {
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
