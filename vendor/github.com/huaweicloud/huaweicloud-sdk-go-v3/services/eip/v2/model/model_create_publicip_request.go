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

// Request Object
type CreatePublicipRequest struct {
	Body *CreatePublicipRequestBody `json:"body,omitempty"`
}

func (o CreatePublicipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePublicipRequest struct{}"
	}

	return strings.Join([]string{"CreatePublicipRequest", string(data)}, " ")
}
