// +build !linux

package net

// VirtualInterfaces returns virtual network interfaces existing in the system.
func VirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool)

	return cardVirtual, nil
}
