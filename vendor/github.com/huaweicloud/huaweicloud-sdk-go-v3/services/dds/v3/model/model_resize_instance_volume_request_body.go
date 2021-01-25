/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ResizeInstanceVolumeRequestBody struct {
	Volume *ResizeInstanceVolumeOption `json:"volume"`
}

func (o ResizeInstanceVolumeRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceVolumeRequestBody struct{}"
	}

	return strings.Join([]string{"ResizeInstanceVolumeRequestBody", string(data)}, " ")
}
