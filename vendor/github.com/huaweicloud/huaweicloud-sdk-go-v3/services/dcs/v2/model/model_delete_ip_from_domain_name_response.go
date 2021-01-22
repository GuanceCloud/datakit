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
type DeleteIpFromDomainNameResponse struct {
	// 域名摘除ip的任务ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteIpFromDomainNameResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteIpFromDomainNameResponse struct{}"
	}

	return strings.Join([]string{"DeleteIpFromDomainNameResponse", string(data)}, " ")
}
