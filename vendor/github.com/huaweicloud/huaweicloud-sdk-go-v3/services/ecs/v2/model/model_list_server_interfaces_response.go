/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListServerInterfacesResponse struct {
	// 云服务器网卡信息列表
	InterfaceAttachments *[]InterfaceAttachment `json:"interfaceAttachments,omitempty"`
	HttpStatusCode       int                    `json:"-"`
}

func (o ListServerInterfacesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServerInterfacesResponse struct{}"
	}

	return strings.Join([]string{"ListServerInterfacesResponse", string(data)}, " ")
}
