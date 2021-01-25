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
type NovaDeleteKeypairResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NovaDeleteKeypairResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaDeleteKeypairResponse struct{}"
	}

	return strings.Join([]string{"NovaDeleteKeypairResponse", string(data)}, " ")
}
