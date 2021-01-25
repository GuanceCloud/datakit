/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TemplateInfo struct {
	// 模板id
	Id *string `json:"id,omitempty"`
	// 模板名
	Title *string `json:"title,omitempty"`
	// 模板描述
	Description *string `json:"description,omitempty"`
	// 模板关联的所有云服务（产品短名）
	Productshorts *[]string `json:"productshorts,omitempty"`
	// 模板关联的云产品
	Products *[]TemplateProductExt `json:"products,omitempty"`
	// 模板标签
	Topic *[]TopicCategory `json:"topic,omitempty"`
	// 模板创建者id
	CreatorId *string `json:"creator_id,omitempty"`
	// 模板创建者,有别名返回别名
	Creator *string `json:"creator,omitempty"`
	// 模板创建者,有别名返回别名
	Nickname *string `json:"nickname,omitempty"`
	// 模板评分（点赞数）
	Score *int32 `json:"score,omitempty"`
	// 模板标签（new、hot等）
	Label *string `json:"label,omitempty"`
	// 代码存储位置
	Store *int32 `json:"store,omitempty"`
	// 获取代码模版所需的信息
	StoreInfo *string `json:"store_info,omitempty"`
	// 模板状态（0:审核中 1: 已上架 2: 未上架（已下架）3: 未上架（合规检查不通过）4：未上架（待上架）5：已删除）
	Status *int32 `json:"status,omitempty"`
	// 访问量
	ViewCount *int32 `json:"view_count,omitempty"`
	// 引用量
	UsageCount *int32 `json:"usage_count,omitempty"`
	// 创建时间
	CreatedAt *string `json:"created_at,omitempty"`
	// 更新时间
	UpdatedAt *string `json:"updated_at,omitempty"`
	// 模板上架时间
	PublishedAt *string `json:"published_at,omitempty"`
	// 点赞状态(1：点赞，0：未点赞)
	FavoriteState *int32 `json:"favorite_state,omitempty"`
	// 模板标签
	Tags *[]TagInfo `json:"tags,omitempty"`
	// 模板类型（0:doc、1:code、2:pipeline、3:devops四种）
	Type *int32 `json:"type,omitempty"`
	// 动、静态代码模板标识（0：动态模板codetemplate，1：静态模板codesample）
	IsStatic *int32 `json:"is_static,omitempty"`
	// 模板相关联的所有维护人账号名称
	Maintainers      *[]string             `json:"maintainers,omitempty"`
	PipelineTemplate *PipelineTemplateInfo `json:"pipeline_template,omitempty"`
	// 平台来源（0:codelabs、1:devstar）
	PlatformSource *int32 `json:"platform_source,omitempty"`
	// 相关文档，示例，帖子
	References *[]Reference `json:"references,omitempty"`
	// 模板自定义参数列表
	Properties *interface{} `json:"properties,omitempty"`
	// dependency信息
	Dependencies *[]interface{} `json:"dependencies,omitempty"`
	// dependency类型
	DependencyType *string `json:"dependency_type,omitempty"`
	// 关联论坛板块id
	ForumId *int32 `json:"forum_id,omitempty"`
	// 模板文件解压缩之后的大小(单位:KB)
	FileSize *int32 `json:"file_size,omitempty"`
	// 部署信息
	Deployment *interface{} `json:"deployment,omitempty"`
	// 模板关联更新态Id
	UpdateId *string `json:"update_id,omitempty"`
	// 是否支持使用CloudIDE运行源码
	IsSupportCloudide *bool `json:"is_support_cloudide,omitempty"`
}

func (o TemplateInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateInfo struct{}"
	}

	return strings.Join([]string{"TemplateInfo", string(data)}, " ")
}
