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

type TemplateJobInfo struct {
	// 应用名称
	ApplicationName string `json:"application_name"`
	// 0 - 将生成的应用代码存储于 repo_info 指定的 CodeHub 仓库中。1 - 将生成的应用代码存储到华为云，任务创建人可以通过 ExportApplicationCode 下载代码压缩包
	RepoType TemplateJobInfoRepoType `json:"repo_type"`
	// Devstar 模板 ID，通过 [模板列表查询接口](https://apiexplorer.developer.huaweicloud.com/apiexplorer/doc?product=DevStar&api=ListPublishedTemplates) 获取相应模板 ID
	TemplateId string `json:"template_id"`
	// 模板的动态参数, 可以从 [模板详情查询接口](https://apiexplorer.developer.huaweicloud.com/apiexplorer/doc?product=DevStar&api=ShowTemplateDetail) 获取
	Properties map[string]string `json:"properties,omitempty"`
	RepoInfo   *RepositoryInfo   `json:"repo_info,omitempty"`
}

func (o TemplateJobInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateJobInfo struct{}"
	}

	return strings.Join([]string{"TemplateJobInfo", string(data)}, " ")
}

type TemplateJobInfoRepoType struct {
	value int32
}

type TemplateJobInfoRepoTypeEnum struct {
	E_0 TemplateJobInfoRepoType
	E_1 TemplateJobInfoRepoType
}

func GetTemplateJobInfoRepoTypeEnum() TemplateJobInfoRepoTypeEnum {
	return TemplateJobInfoRepoTypeEnum{
		E_0: TemplateJobInfoRepoType{
			value: 0,
		}, E_1: TemplateJobInfoRepoType{
			value: 1,
		},
	}
}

func (c TemplateJobInfoRepoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *TemplateJobInfoRepoType) UnmarshalJSON(b []byte) error {
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
