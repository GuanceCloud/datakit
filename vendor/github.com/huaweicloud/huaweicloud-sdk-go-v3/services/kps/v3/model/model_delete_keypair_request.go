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
type DeleteKeypairRequest struct {
	KeypairName string `json:"keypair_name"`
}

func (o DeleteKeypairRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteKeypairRequest struct{}"
	}

	return strings.Join([]string{"DeleteKeypairRequest", string(data)}, " ")
}
