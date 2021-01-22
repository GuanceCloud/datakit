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

// Request Object
type ListProjectsV4Request struct {
	Offset      int32   `json:"offset"`
	Limit       int32   `json:"limit"`
	Search      *string `json:"search,omitempty"`
	ProjectType *string `json:"project_type,omitempty"`
	Sort        *string `json:"sort,omitempty"`
	Archive     *string `json:"archive,omitempty"`
	QueryType   *string `json:"query_type,omitempty"`
}

func (o ListProjectsV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectsV4Request struct{}"
	}

	return strings.Join([]string{"ListProjectsV4Request", string(data)}, " ")
}
