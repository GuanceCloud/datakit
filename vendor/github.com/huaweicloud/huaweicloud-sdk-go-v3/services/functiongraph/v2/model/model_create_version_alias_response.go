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

// Response Object
type CreateVersionAliasResponse struct {
	// 要获取的别名名称。
	Name *string `json:"name,omitempty"`
	// 别名对应的版本名称。
	Version *string `json:"version,omitempty"`
	// 别名描述信息。
	Description *string `json:"description,omitempty"`
	// 别名最后修改时间。
	LastModified *sdktime.SdkTime `json:"last_modified,omitempty"`
	// 版本别名唯一标识。
	AliasUrn *string `json:"alias_urn,omitempty"`
	// 灰度版本信息
	AdditionalVersionWeights map[string]int32 `json:"additional_version_weights,omitempty"`
	HttpStatusCode           int              `json:"-"`
}

func (o CreateVersionAliasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVersionAliasResponse struct{}"
	}

	return strings.Join([]string{"CreateVersionAliasResponse", string(data)}, " ")
}
