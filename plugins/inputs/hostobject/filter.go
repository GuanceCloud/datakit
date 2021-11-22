package hostobject

func DiskIgnoreFs(ignoreFs []string) (map[string]bool, error) {
	ignoreFsMap := map[string]bool{}
	for _, fs := range ignoreFs {
		ignoreFsMap[fs] = true
	}
	return ignoreFsMap, nil
}

func NetIgnoreIfaces() (map[string]bool, error) {
	return NetVirtualInterfaces()
}
