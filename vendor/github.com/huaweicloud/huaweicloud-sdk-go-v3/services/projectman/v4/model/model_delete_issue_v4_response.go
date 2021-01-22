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
type DeleteIssueV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteIssueV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteIssueV4Response struct{}"
	}

	return strings.Join([]string{"DeleteIssueV4Response", string(data)}, " ")
}
