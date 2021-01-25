/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateBaremetalServerMetadataResponse struct {
	Metadata       *KeyValue `json:"metadata,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o UpdateBaremetalServerMetadataResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateBaremetalServerMetadataResponse struct{}"
	}

	return strings.Join([]string{"UpdateBaremetalServerMetadataResponse", string(data)}, " ")
}
