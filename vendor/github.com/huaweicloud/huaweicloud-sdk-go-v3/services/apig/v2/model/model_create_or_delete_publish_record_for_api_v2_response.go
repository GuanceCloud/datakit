/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

// Response Object
type CreateOrDeletePublishRecordForApiV2Response struct {
	// 发布记录的ID
	PublishId *string `json:"publish_id,omitempty"`
	// API编号
	ApiId *string `json:"api_id,omitempty"`
	// API名称
	ApiName *string `json:"api_name,omitempty"`
	// 发布的环境编号
	EnvId *string `json:"env_id,omitempty"`
	// 发布描述
	Remark *string `json:"remark,omitempty"`
	// 发布时间
	PublishTime *sdktime.SdkTime `json:"publish_time,omitempty"`
	// 在线的版本号
	VersionId      *string `json:"version_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateOrDeletePublishRecordForApiV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateOrDeletePublishRecordForApiV2Response struct{}"
	}

	return strings.Join([]string{"CreateOrDeletePublishRecordForApiV2Response", string(data)}, " ")
}
