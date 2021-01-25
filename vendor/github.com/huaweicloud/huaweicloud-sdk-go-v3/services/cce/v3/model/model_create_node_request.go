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
type CreateNodeRequest struct {
	ClusterId       string                            `json:"cluster_id"`
	ContentType     string                            `json:"Content-Type"`
	NodepoolScaleUp *CreateNodeRequestNodepoolScaleUp `json:"nodepoolScaleUp,omitempty"`
	Body            *V3NodeCreateRequest              `json:"body,omitempty"`
}

func (o CreateNodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNodeRequest struct{}"
	}

	return strings.Join([]string{"CreateNodeRequest", string(data)}, " ")
}

type CreateNodeRequestNodepoolScaleUp struct {
	value string
}

type CreateNodeRequestNodepoolScaleUpEnum struct {
	NODEPOOL_SCALE_UP CreateNodeRequestNodepoolScaleUp
}

func GetCreateNodeRequestNodepoolScaleUpEnum() CreateNodeRequestNodepoolScaleUpEnum {
	return CreateNodeRequestNodepoolScaleUpEnum{
		NODEPOOL_SCALE_UP: CreateNodeRequestNodepoolScaleUp{
			value: "NodepoolScaleUp",
		},
	}
}

func (c CreateNodeRequestNodepoolScaleUp) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateNodeRequestNodepoolScaleUp) UnmarshalJSON(b []byte) error {
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
