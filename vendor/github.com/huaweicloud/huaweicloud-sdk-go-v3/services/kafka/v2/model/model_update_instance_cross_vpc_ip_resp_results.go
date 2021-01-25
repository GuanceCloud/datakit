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

// 修改broker跨VPC访问的结果。
type UpdateInstanceCrossVpcIpRespResults struct {
	// advertised.listeners IP/域名。
	AdvertisedIp *string `json:"advertised_ip,omitempty"`
	// 修改broker跨VPC访问的状态。
	Success *string `json:"success,omitempty"`
	// listeners IP。
	Ip *string `json:"ip,omitempty"`
}

func (o UpdateInstanceCrossVpcIpRespResults) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceCrossVpcIpRespResults struct{}"
	}

	return strings.Join([]string{"UpdateInstanceCrossVpcIpRespResults", string(data)}, " ")
}
