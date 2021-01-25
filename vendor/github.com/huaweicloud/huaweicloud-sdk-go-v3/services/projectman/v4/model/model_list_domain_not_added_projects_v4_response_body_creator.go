/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 创建者信息
type ListDomainNotAddedProjectsV4ResponseBodyCreator struct {
	// 创建人numId
	UserNumId *int32 `json:"user_num_id,omitempty"`
	// 创建人id
	UserId *string `json:"user_id,omitempty"`
	// 创建人姓名
	UserName *string `json:"user_name,omitempty"`
	// 创建人租户id
	DomainId *string `json:"domain_id,omitempty"`
	// 创建人租户名称
	DomainName *string `json:"domain_name,omitempty"`
	// 创建人租户昵称
	NickName *string `json:"nick_name,omitempty"`
}

func (o ListDomainNotAddedProjectsV4ResponseBodyCreator) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainNotAddedProjectsV4ResponseBodyCreator struct{}"
	}

	return strings.Join([]string{"ListDomainNotAddedProjectsV4ResponseBodyCreator", string(data)}, " ")
}
