/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
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
type ListInstanceSnapshotsRequest struct {
	ApplicationId   string                             `json:"application_id"`
	ComponentId     string                             `json:"component_id"`
	InstanceId      string                             `json:"instance_id"`
	Limit           *int32                             `json:"limit,omitempty"`
	Offset          *int32                             `json:"offset,omitempty"`
	SnapshotOrderBy *string                            `json:"snapshot_order_by,omitempty"`
	Order           *ListInstanceSnapshotsRequestOrder `json:"order,omitempty"`
}

func (o ListInstanceSnapshotsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstanceSnapshotsRequest struct{}"
	}

	return strings.Join([]string{"ListInstanceSnapshotsRequest", string(data)}, " ")
}

type ListInstanceSnapshotsRequestOrder struct {
	value string
}

type ListInstanceSnapshotsRequestOrderEnum struct {
	DESC ListInstanceSnapshotsRequestOrder
	ASC  ListInstanceSnapshotsRequestOrder
}

func GetListInstanceSnapshotsRequestOrderEnum() ListInstanceSnapshotsRequestOrderEnum {
	return ListInstanceSnapshotsRequestOrderEnum{
		DESC: ListInstanceSnapshotsRequestOrder{
			value: "desc",
		},
		ASC: ListInstanceSnapshotsRequestOrder{
			value: "asc",
		},
	}
}

func (c ListInstanceSnapshotsRequestOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstanceSnapshotsRequestOrder) UnmarshalJSON(b []byte) error {
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
