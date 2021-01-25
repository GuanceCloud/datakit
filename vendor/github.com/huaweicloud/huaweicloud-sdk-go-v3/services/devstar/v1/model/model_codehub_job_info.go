/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type CodehubJobInfo struct {
	// 应用名称
	ApplicationName string `json:"application_name"`
	// 代码存放的ssh地址
	CodeUrl string `json:"code_url"`
	// CodeHub 仓库所在的 Region ID。华南-广州: cn-south-1, 华东-上海二：cn-east-2，华北-北京一： cn-north-1 ，华北-北京四：cn-north-4
	RegionId string `json:"region_id"`
	// 0 - 将生成的应用代码存储于 repo_info 指定的 CodeHub 仓库中。1 - 将生成的应用代码存储到华为云，任务创建人可以通过 ExportApplicationCode 下载代码压缩包
	RepoType CodehubJobInfoRepoType `json:"repo_type"`
	// 可以根据 template-metadata.json 获取动态参数 ID 以及规则
	Properties map[string]string `json:"properties,omitempty"`
	RepoInfo   *RepositoryInfo   `json:"repo_info,omitempty"`
}

func (o CodehubJobInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CodehubJobInfo struct{}"
	}

	return strings.Join([]string{"CodehubJobInfo", string(data)}, " ")
}

type CodehubJobInfoRepoType struct {
	value int32
}

type CodehubJobInfoRepoTypeEnum struct {
	E_0 CodehubJobInfoRepoType
	E_1 CodehubJobInfoRepoType
}

func GetCodehubJobInfoRepoTypeEnum() CodehubJobInfoRepoTypeEnum {
	return CodehubJobInfoRepoTypeEnum{
		E_0: CodehubJobInfoRepoType{
			value: 0,
		}, E_1: CodehubJobInfoRepoType{
			value: 1,
		},
	}
}

func (c CodehubJobInfoRepoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CodehubJobInfoRepoType) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}
