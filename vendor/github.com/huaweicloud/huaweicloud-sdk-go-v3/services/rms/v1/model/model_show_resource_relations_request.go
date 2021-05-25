/*
 * RMS
 *
 * Resource Manager Api
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
type ShowResourceRelationsRequest struct {
	ResourceId string                                `json:"resource_id"`
	Direction  ShowResourceRelationsRequestDirection `json:"direction"`
	Limit      *int32                                `json:"limit,omitempty"`
	Marker     *string                               `json:"marker,omitempty"`
}

func (o ShowResourceRelationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceRelationsRequest struct{}"
	}

	return strings.Join([]string{"ShowResourceRelationsRequest", string(data)}, " ")
}

type ShowResourceRelationsRequestDirection struct {
	value string
}

type ShowResourceRelationsRequestDirectionEnum struct {
	IN  ShowResourceRelationsRequestDirection
	OUT ShowResourceRelationsRequestDirection
}

func GetShowResourceRelationsRequestDirectionEnum() ShowResourceRelationsRequestDirectionEnum {
	return ShowResourceRelationsRequestDirectionEnum{
		IN: ShowResourceRelationsRequestDirection{
			value: "in",
		},
		OUT: ShowResourceRelationsRequestDirection{
			value: "out",
		},
	}
}

func (c ShowResourceRelationsRequestDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowResourceRelationsRequestDirection) UnmarshalJSON(b []byte) error {
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
