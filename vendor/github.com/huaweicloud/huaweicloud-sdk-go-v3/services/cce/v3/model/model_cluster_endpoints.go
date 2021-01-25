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

type ClusterEndpoints struct {
	// 集群访问地址的类型 - Internal：用户子网内访问的地址 - External：公网访问的地址
	Type *string `json:"type,omitempty"`
	// 集群中 kube-apiserver 的访问地址
	Url *string `json:"url,omitempty"`
}

func (o ClusterEndpoints) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClusterEndpoints struct{}"
	}

	return strings.Join([]string{"ClusterEndpoints", string(data)}, " ")
}
