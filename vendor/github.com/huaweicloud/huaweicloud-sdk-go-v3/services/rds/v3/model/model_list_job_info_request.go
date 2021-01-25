/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListJobInfoRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	Id        string  `json:"id"`
}

func (o ListJobInfoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListJobInfoRequest struct{}"
	}

	return strings.Join([]string{"ListJobInfoRequest", string(data)}, " ")
}
