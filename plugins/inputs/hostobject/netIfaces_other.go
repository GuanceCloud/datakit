// +build !linux

package hostobject

// returns virtual network interfaces existing in the system.
func NetVirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool, 0)
	return cardVirtual, nil
}
