/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListScalingTagInfosByResourceIdResponse struct {
	// 资源标签列表。
	Tags *[]TagsSingleValue `json:"tags,omitempty"`
	// 系统资源标签列表。
	SysTags        *[]TagsSingleValue `json:"sys_tags,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListScalingTagInfosByResourceIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListScalingTagInfosByResourceIdResponse struct{}"
	}

	return strings.Join([]string{"ListScalingTagInfosByResourceIdResponse", string(data)}, " ")
}
