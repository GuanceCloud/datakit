// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
