/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

type UpdateVersionAliasRequestBody struct {
	// 要获取的别名名称。
	Name string `json:"name"`
	// 别名对应的版本名称。
	Version string `json:"version"`
	// 别名描述信息。
	Description *string `json:"description,omitempty"`
	// 别名最后修改时间。
	LastModified *sdktime.SdkTime `json:"last_modified"`
	// 版本别名唯一标识。
	AliasUrn string `json:"alias_urn"`
	// 灰度版本信息
	AdditionalVersionWeights map[string]int32 `json:"additional_version_weights,omitempty"`
}

func (o UpdateVersionAliasRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVersionAliasRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateVersionAliasRequestBody", string(data)}, " ")
}
