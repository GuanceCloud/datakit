/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type PersistentVolumeClaim struct {
	// API版本，固定值**v1**
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值**PersistentVolumeClaim**
	Kind     string                         `json:"kind"`
	Metadata *PersistentVolumeClaimMetadata `json:"metadata"`
	Spec     *PersistentVolumeClaimSpec     `json:"spec"`
	Status   *PersistentVolumeClaimStatus   `json:"status,omitempty"`
}

func (o PersistentVolumeClaim) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PersistentVolumeClaim struct{}"
	}

	return strings.Join([]string{"PersistentVolumeClaim", string(data)}, " ")
}
