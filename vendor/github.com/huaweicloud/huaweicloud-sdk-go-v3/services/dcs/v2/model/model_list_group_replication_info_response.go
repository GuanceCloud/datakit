/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListGroupReplicationInfoResponse struct {
	// 分片列表
	GroupList *[]InstanceGroupListInfo `json:"group_list,omitempty"`
	// 实例分片总数。
	GroupCount     *int32 `json:"group_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListGroupReplicationInfoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListGroupReplicationInfoResponse struct{}"
	}

	return strings.Join([]string{"ListGroupReplicationInfoResponse", string(data)}, " ")
}
