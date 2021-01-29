/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type DnsAssignMent struct {
	// 端口hostname
	Hostname *string `json:"hostname,omitempty"`
	// 端口IP地址
	IpAddress *string `json:"ip_address,omitempty"`
	// 端口内网fqdn
	Fqdn *string `json:"fqdn,omitempty"`
}

func (o DnsAssignMent) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DnsAssignMent struct{}"
	}

	return strings.Join([]string{"DnsAssignMent", string(data)}, " ")
}
