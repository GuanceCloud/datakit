/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListImageAutoSyncReposDetailsResponse struct {
	// 镜像自动同步规则
	Body           *[]SyncRepo `json:"body,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ListImageAutoSyncReposDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListImageAutoSyncReposDetailsResponse struct{}"
	}

	return strings.Join([]string{"ListImageAutoSyncReposDetailsResponse", string(data)}, " ")
}
