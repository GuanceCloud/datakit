// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package hostobject

import (
	"fmt"
	"os"
)

const vnicDevPath = "/sys/devices/virtual/net/"

// NetVirtualInterfaces returns virtual network card existing in the system.
func NetVirtualInterfaces() (map[string]bool, error) {
	cardVirtual := make(map[string]bool)

	v, err := os.ReadDir(vnicDevPath)
	if err != nil {
		return nil, fmt.Errorf("read dir %s` failed: %w",
			vnicDevPath, err)
	}

	for _, v := range v {
		if v.IsDir() {
			cardVirtual[v.Name()] = true
		}
	}

	return cardVirtual, nil
}
