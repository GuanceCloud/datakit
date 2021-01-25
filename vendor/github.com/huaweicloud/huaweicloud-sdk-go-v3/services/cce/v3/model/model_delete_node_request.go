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

// Request Object
type DeleteNodeRequest struct {
	ClusterId         string                              `json:"cluster_id"`
	NodeId            string                              `json:"node_id"`
	ContentType       string                              `json:"Content-Type"`
	ErrorStatus       *string                             `json:"errorStatus,omitempty"`
	NodepoolScaleDown *DeleteNodeRequestNodepoolScaleDown `json:"nodepoolScaleDown,omitempty"`
}

func (o DeleteNodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNodeRequest struct{}"
	}

	return strings.Join([]string{"DeleteNodeRequest", string(data)}, " ")
}

type DeleteNodeRequestNodepoolScaleDown struct {
	value string
}

type DeleteNodeRequestNodepoolScaleDownEnum struct {
	NO_SCALE_DOWN DeleteNodeRequestNodepoolScaleDown
}

func GetDeleteNodeRequestNodepoolScaleDownEnum() DeleteNodeRequestNodepoolScaleDownEnum {
	return DeleteNodeRequestNodepoolScaleDownEnum{
		NO_SCALE_DOWN: DeleteNodeRequestNodepoolScaleDown{
			value: "NoScaleDown",
		},
	}
}

func (c DeleteNodeRequestNodepoolScaleDown) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteNodeRequestNodepoolScaleDown) UnmarshalJSON(b []byte) error {
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
