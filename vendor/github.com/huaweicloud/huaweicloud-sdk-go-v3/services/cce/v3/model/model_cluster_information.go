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

type ClusterInformation struct {
	Spec *ClusterInformationSpec `json:"spec"`
}

func (o ClusterInformation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClusterInformation struct{}"
	}

	return strings.Join([]string{"ClusterInformation", string(data)}, " ")
}
