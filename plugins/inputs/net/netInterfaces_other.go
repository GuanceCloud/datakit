// +build !linux

package net

// returns virtual network interfaces existing in the system.
func VirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool, 0)

	return cardVirtual, nil
}
