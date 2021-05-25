// +build !windows
// +build go1.7

package luser

import "strconv"

func lookupGroup(name string) (*Group, error) {
	if _, err := strconv.Atoi(name); err == nil {
		return nil, UnknownGroupError(name)
	}

	if dscacheutilExe != "" {
		g, err := dsGroup(name)
		if err == nil {
			return lgroup(g), nil
		}
	}

	g, err := getentGroup(name)
	if err == nil {
		return lgroup(g), nil
	}

	return nil, UnknownGroupError(name)
}

func lookupGroupId(gid string) (*Group, error) {
	id, err := strconv.Atoi(gid)
	if err != nil {
		return nil, err
	}

	if dscacheutilExe != "" {
		g, err := dsGroupId(gid)
		if err == nil {
			return lgroup(g), nil
		}
	}

	g, err := getentGroup(gid)
	if err == nil {
		return lgroup(g), nil
	}

	return nil, UnknownGroupIdError(id)
}
