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

type CceClusterNodeInformation struct {
	Metadata *CceClusterNodeInformationMetadata `json:"metadata"`
}

func (o CceClusterNodeInformation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CceClusterNodeInformation struct{}"
	}

	return strings.Join([]string{"CceClusterNodeInformation", string(data)}, " ")
}
