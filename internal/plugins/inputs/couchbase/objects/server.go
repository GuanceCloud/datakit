// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

type Server struct {
	Hostname string            `json:"hostname"`
	URI      string            `json:"uri"`
	Stats    map[string]string `json:"stats"`
}

type Servers struct {
	Servers []Server `json:"servers"`
}
