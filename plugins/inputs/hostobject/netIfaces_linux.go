//go:build linux
// +build linux

package hostobject

import (
	"os/exec"
	"strings"
)

// NetVirtualInterfaces returns virtual network card existing in the system.
func NetVirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool)
	var data string
	// mock data
	if len(mockData) == 1 {
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
