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
type AddApplyJoinProjectForAgcResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AddApplyJoinProjectForAgcResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddApplyJoinProjectForAgcResponse struct{}"
	}

	return strings.Join([]string{"AddApplyJoinProjectForAgcResponse", string(data)}, " ")
}
