/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateInstanceCrossVpcIpResponse struct {
	// 修改跨VPC访问结果。
	Success *string `json:"success,omitempty"`
	// 修改broker跨VPC访问的结果列表。
	Results        *[]UpdateInstanceCrossVpcIpRespResults `json:"results,omitempty"`
	HttpStatusCode int                                    `json:"-"`
}

func (o UpdateInstanceCrossVpcIpResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceCrossVpcIpResponse struct{}"
	}

	return strings.Join([]string{"UpdateInstanceCrossVpcIpResponse", string(data)}, " ")
}
