/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListQuotasResponse struct {
	Quotas         *ResourceResp `json:"quotas,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQuotasResponse struct{}"
	}

	return strings.Join([]string{"ListQuotasResponse", string(data)}, " ")
}
