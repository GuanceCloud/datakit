/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreatePostPaidServersRequest struct {
	Body *CreatePostPaidServersRequestBody `json:"body,omitempty"`
}

func (o CreatePostPaidServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePostPaidServersRequest struct{}"
	}

	return strings.Join([]string{"CreatePostPaidServersRequest", string(data)}, " ")
}
