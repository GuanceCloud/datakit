/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type UpdateKeypairDescriptionRequest struct {
	KeypairName string                               `json:"keypair_name"`
	Body        *UpdateKeypairDescriptionRequestBody `json:"body,omitempty"`
}

func (o UpdateKeypairDescriptionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeypairDescriptionRequest struct{}"
	}

	return strings.Join([]string{"UpdateKeypairDescriptionRequest", string(data)}, " ")
}
