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

// Response Object
type NovaCreateKeypairResponse struct {
	Keypair        *NovaKeypair `json:"keypair,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o NovaCreateKeypairResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaCreateKeypairResponse struct{}"
	}

	return strings.Join([]string{"NovaCreateKeypairResponse", string(data)}, " ")
}
