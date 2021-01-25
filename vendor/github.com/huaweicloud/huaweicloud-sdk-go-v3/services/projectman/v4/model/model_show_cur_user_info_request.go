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
type ShowCurUserInfoRequest struct {
}

func (o ShowCurUserInfoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCurUserInfoRequest struct{}"
	}

	return strings.Join([]string{"ShowCurUserInfoRequest", string(data)}, " ")
}
