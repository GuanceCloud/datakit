// +build go1.7

package luser

import (
	"os/user"
	"strings"
)

func getentGroup(key string) (*user.Group, error) {
	entity, err := getent("group", key)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(entity, ":")
	if len(parts) < 3 {
		return nil, errNotFound
	}

	return &user.Group{
		Name: parts[0],
		Gid:  parts[2],
	}, nil
}
