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

type TemplateProductExt struct {
	// 产品id
	Id *string `json:"id,omitempty"`
	// 产品短名
	Productshort *string `json:"productshort,omitempty"`
	// 产品名
	ProductName *string `json:"product_name,omitempty"`
	// 首页链接
	HomeLink *string `json:"home_link,omitempty"`
	// api调试链接
	ApiLink *string `json:"api_link,omitempty"`
	// sdk下载链接
	SdkLink *string `json:"sdk_link,omitempty"`
	// 文档链接
	DocLink *string `json:"doc_link,omitempty"`
	// logo链接
	LogoLink *string `json:"logo_link,omitempty"`
}

func (o TemplateProductExt) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateProductExt struct{}"
	}

	return strings.Join([]string{"TemplateProductExt", string(data)}, " ")
}
