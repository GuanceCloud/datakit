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
type UpdateNickNameV4Request struct {
	Body *UpdateUserNickNameRequestV4 `json:"body,omitempty"`
}

func (o UpdateNickNameV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNickNameV4Request struct{}"
	}

	return strings.Join([]string{"UpdateNickNameV4Request", string(data)}, " ")
}
