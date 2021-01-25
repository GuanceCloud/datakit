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

// Response Object
type BatchDeleteIssuesV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteIssuesV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteIssuesV4Response struct{}"
	}

	return strings.Join([]string{"BatchDeleteIssuesV4Response", string(data)}, " ")
}
