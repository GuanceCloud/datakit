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
type ShowResourceHistoryRequest struct {
	ResourceId         string                                        `json:"resource_id"`
	Marker             *string                                       `json:"marker,omitempty"`
	Limit              *int32                                        `json:"limit,omitempty"`
	EarlierTime        *int64                                        `json:"earlier_time,omitempty"`
	LaterTime          *int64                                        `json:"later_time,omitempty"`
	ChronologicalOrder *ShowResourceHistoryRequestChronologicalOrder `json:"chronological_order,omitempty"`
}

func (o ShowResourceHistoryRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceHistoryRequest struct{}"
	}

	return strings.Join([]string{"ShowResourceHistoryRequest", string(data)}, " ")
}

type ShowResourceHistoryRequestChronologicalOrder struct {
	value string
}

type ShowResourceHistoryRequestChronologicalOrderEnum struct {
	FORWARD ShowResourceHistoryRequestChronologicalOrder
	REVERSE ShowResourceHistoryRequestChronologicalOrder
}

func GetShowResourceHistoryRequestChronologicalOrderEnum() ShowResourceHistoryRequestChronologicalOrderEnum {
	return ShowResourceHistoryRequestChronologicalOrderEnum{
		FORWARD: ShowResourceHistoryRequestChronologicalOrder{
			value: "Forward",
		},
		REVERSE: ShowResourceHistoryRequestChronologicalOrder{
			value: "Reverse",
		},
	}
}

func (c ShowResourceHistoryRequestChronologicalOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowResourceHistoryRequestChronologicalOrder) UnmarshalJSON(b []byte) error {
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
