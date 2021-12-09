//go:build !linux
// +build !linux

package hostobject

// NetVirtualInterfaces returns virtual network interfaces existing in the system.
func NetVirtualInterfaces(mockData ...string) (map[string]bool, error) {
	cardVirtual := make(map[string]bool)
	return cardVirtual, nil
}
