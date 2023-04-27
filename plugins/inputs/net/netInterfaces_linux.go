// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package net

import (
	"os/exec"
	"strings"
)

// VirtualInterfaces returns virtual network card existing in the system.
func VirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool)
	var data string
	// mock data
	if len(mockData) > 0 {
		data = mockData[0]
	} else {
		b, err := exec.Command("ls", "/sys/devices/virtual/net/").CombinedOutput()
		if err != nil {
			return nil, err
		}
		data = string(b)
	}

	for _, v := range strings.Split(data, "\n") {
		if len(v) > 0 {
			cardVirtual[v] = true
		}
	}

	return cardVirtual, nil
}
