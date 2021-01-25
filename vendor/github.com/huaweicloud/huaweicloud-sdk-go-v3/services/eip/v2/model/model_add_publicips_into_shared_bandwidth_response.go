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

// Response Object
type AddPublicipsIntoSharedBandwidthResponse struct {
	Bandwidth      *BandwidthRespInsert `json:"bandwidth,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o AddPublicipsIntoSharedBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddPublicipsIntoSharedBandwidthResponse struct{}"
	}

	return strings.Join([]string{"AddPublicipsIntoSharedBandwidthResponse", string(data)}, " ")
}
