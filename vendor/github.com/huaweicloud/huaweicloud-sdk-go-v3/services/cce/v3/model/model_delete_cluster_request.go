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
type DeleteClusterRequest struct {
	ClusterId   string                         `json:"cluster_id"`
	ContentType string                         `json:"Content-Type"`
	ErrorStatus *string                        `json:"errorStatus,omitempty"`
	DeleteEfs   *DeleteClusterRequestDeleteEfs `json:"delete_efs,omitempty"`
	DeleteEni   *DeleteClusterRequestDeleteEni `json:"delete_eni,omitempty"`
	DeleteEvs   *DeleteClusterRequestDeleteEvs `json:"delete_evs,omitempty"`
	DeleteNet   *DeleteClusterRequestDeleteNet `json:"delete_net,omitempty"`
	DeleteObs   *DeleteClusterRequestDeleteObs `json:"delete_obs,omitempty"`
	DeleteSfs   *DeleteClusterRequestDeleteSfs `json:"delete_sfs,omitempty"`
}

func (o DeleteClusterRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteClusterRequest struct{}"
	}

	return strings.Join([]string{"DeleteClusterRequest", string(data)}, " ")
}

type DeleteClusterRequestDeleteEfs struct {
	value string
}

type DeleteClusterRequestDeleteEfsEnum struct {
	TRUE  DeleteClusterRequestDeleteEfs
	BLOCK DeleteClusterRequestDeleteEfs
	TRY   DeleteClusterRequestDeleteEfs
	FALSE DeleteClusterRequestDeleteEfs
	SKIP  DeleteClusterRequestDeleteEfs
}

func GetDeleteClusterRequestDeleteEfsEnum() DeleteClusterRequestDeleteEfsEnum {
	return DeleteClusterRequestDeleteEfsEnum{
		TRUE: DeleteClusterRequestDeleteEfs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEfs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEfs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEfs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEfs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEfs) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEfs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteEni struct {
	value string
}

type DeleteClusterRequestDeleteEniEnum struct {
	TRUE  DeleteClusterRequestDeleteEni
	BLOCK DeleteClusterRequestDeleteEni
	TRY   DeleteClusterRequestDeleteEni
	FALSE DeleteClusterRequestDeleteEni
	SKIP  DeleteClusterRequestDeleteEni
}

func GetDeleteClusterRequestDeleteEniEnum() DeleteClusterRequestDeleteEniEnum {
	return DeleteClusterRequestDeleteEniEnum{
		TRUE: DeleteClusterRequestDeleteEni{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEni{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEni{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEni{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEni{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEni) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEni) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteEvs struct {
	value string
}

type DeleteClusterRequestDeleteEvsEnum struct {
	TRUE  DeleteClusterRequestDeleteEvs
	BLOCK DeleteClusterRequestDeleteEvs
	TRY   DeleteClusterRequestDeleteEvs
	FALSE DeleteClusterRequestDeleteEvs
	SKIP  DeleteClusterRequestDeleteEvs
}

func GetDeleteClusterRequestDeleteEvsEnum() DeleteClusterRequestDeleteEvsEnum {
	return DeleteClusterRequestDeleteEvsEnum{
		TRUE: DeleteClusterRequestDeleteEvs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEvs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEvs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEvs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEvs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEvs) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEvs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteNet struct {
	value string
}

type DeleteClusterRequestDeleteNetEnum struct {
	TRUE  DeleteClusterRequestDeleteNet
	BLOCK DeleteClusterRequestDeleteNet
	TRY   DeleteClusterRequestDeleteNet
	FALSE DeleteClusterRequestDeleteNet
	SKIP  DeleteClusterRequestDeleteNet
}

func GetDeleteClusterRequestDeleteNetEnum() DeleteClusterRequestDeleteNetEnum {
	return DeleteClusterRequestDeleteNetEnum{
		TRUE: DeleteClusterRequestDeleteNet{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteNet{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteNet{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteNet{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteNet{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteNet) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteNet) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteObs struct {
	value string
}

type DeleteClusterRequestDeleteObsEnum struct {
	TRUE  DeleteClusterRequestDeleteObs
	BLOCK DeleteClusterRequestDeleteObs
	TRY   DeleteClusterRequestDeleteObs
	FALSE DeleteClusterRequestDeleteObs
	SKIP  DeleteClusterRequestDeleteObs
}

func GetDeleteClusterRequestDeleteObsEnum() DeleteClusterRequestDeleteObsEnum {
	return DeleteClusterRequestDeleteObsEnum{
		TRUE: DeleteClusterRequestDeleteObs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteObs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteObs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteObs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteObs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteObs) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteObs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteSfs struct {
	value string
}

type DeleteClusterRequestDeleteSfsEnum struct {
	TRUE  DeleteClusterRequestDeleteSfs
	BLOCK DeleteClusterRequestDeleteSfs
	TRY   DeleteClusterRequestDeleteSfs
	FALSE DeleteClusterRequestDeleteSfs
	SKIP  DeleteClusterRequestDeleteSfs
}

func GetDeleteClusterRequestDeleteSfsEnum() DeleteClusterRequestDeleteSfsEnum {
	return DeleteClusterRequestDeleteSfsEnum{
		TRUE: DeleteClusterRequestDeleteSfs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteSfs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteSfs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteSfs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteSfs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteSfs) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteSfs) UnmarshalJSON(b []byte) error {
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
