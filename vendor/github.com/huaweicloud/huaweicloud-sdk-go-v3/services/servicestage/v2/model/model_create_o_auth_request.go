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
type CreateOAuthRequest struct {
	RepoType CreateOAuthRequestRepoType `json:"repo_type"`
	Tag      *string                    `json:"tag,omitempty"`
	Body     *OAuth                     `json:"body,omitempty"`
}

func (o CreateOAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateOAuthRequest struct{}"
	}

	return strings.Join([]string{"CreateOAuthRequest", string(data)}, " ")
}

type CreateOAuthRequestRepoType struct {
	value string
}

type CreateOAuthRequestRepoTypeEnum struct {
	GITHUB    CreateOAuthRequestRepoType
	GITLAB    CreateOAuthRequestRepoType
	GITEE     CreateOAuthRequestRepoType
	BITBUCKET CreateOAuthRequestRepoType
}

func GetCreateOAuthRequestRepoTypeEnum() CreateOAuthRequestRepoTypeEnum {
	return CreateOAuthRequestRepoTypeEnum{
		GITHUB: CreateOAuthRequestRepoType{
			value: "github",
		},
		GITLAB: CreateOAuthRequestRepoType{
			value: "gitlab",
		},
		GITEE: CreateOAuthRequestRepoType{
			value: "gitee",
		},
		BITBUCKET: CreateOAuthRequestRepoType{
			value: "bitbucket",
		},
	}
}

func (c CreateOAuthRequestRepoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateOAuthRequestRepoType) UnmarshalJSON(b []byte) error {
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
