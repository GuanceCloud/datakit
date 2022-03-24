package hostobject

func NetIgnoreIfaces() (map[string]bool, error) {
	return NetVirtualInterfaces()
}
