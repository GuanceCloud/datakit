/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListApplicationsResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 返回的Application个数。该参数不受offset和limit影响，即返回的是您账户下所有的Application个数。
	ApplicationCount *int32             `json:"application_count,omitempty"`
	Applications     *[]ApplicationItem `json:"applications,omitempty"`
	HttpStatusCode   int                `json:"-"`
}

func (o ListApplicationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationsResponse struct{}"
	}

	return strings.Join([]string{"ListApplicationsResponse", string(data)}, " ")
}
