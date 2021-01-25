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
type ListProjectTagsRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
}

func (o ListProjectTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectTagsRequest struct{}"
	}

	return strings.Join([]string{"ListProjectTagsRequest", string(data)}, " ")
}
