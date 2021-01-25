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
type CreatePasswordAuthRequest struct {
	RepoType CreatePasswordAuthRequestRepoType `json:"repo_type"`
	Body     *AccessPassword                   `json:"body,omitempty"`
}

func (o CreatePasswordAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePasswordAuthRequest struct{}"
	}

	return strings.Join([]string{"CreatePasswordAuthRequest", string(data)}, " ")
}

type CreatePasswordAuthRequestRepoType struct {
	value string
}

type CreatePasswordAuthRequestRepoTypeEnum struct {
	GITHUB    CreatePasswordAuthRequestRepoType
	DEVCLOUD  CreatePasswordAuthRequestRepoType
	BITBUCKET CreatePasswordAuthRequestRepoType
}

func GetCreatePasswordAuthRequestRepoTypeEnum() CreatePasswordAuthRequestRepoTypeEnum {
	return CreatePasswordAuthRequestRepoTypeEnum{
		GITHUB: CreatePasswordAuthRequestRepoType{
			value: "github",
		},
		DEVCLOUD: CreatePasswordAuthRequestRepoType{
			value: "devcloud",
		},
		BITBUCKET: CreatePasswordAuthRequestRepoType{
			value: "bitbucket",
		},
	}
}

func (c CreatePasswordAuthRequestRepoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePasswordAuthRequestRepoType) UnmarshalJSON(b []byte) error {
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
