// +build go1.7

package luser

import (
	"bufio"
	"bytes"
	"os/user"
	"strings"
)

func dsParseGroup(data []byte) *user.Group {
	g := &user.Group{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.Index(line, ": "); i >= 0 {
			value := line[i+2:]

			switch line[:i] {
			case "name":
				g.Name = value
			case "gid":
				g.Gid = value
			}
		}
	}

	return g
}

func dsGroup(group string) (*user.Group, error) {
	data, err := command(dscacheutilExe, "-q", "group", "-a", "name", group)
	if err != nil {
		return nil, err
	}

	g := dsParseGroup(data)
	if g.Name != "" && g.Gid != "" {
		return g, nil
	}

	return nil, errNotFound
}

func dsGroupId(group string) (*user.Group, error) {
	data, err := command(dscacheutilExe, "-q", "group", "-a", "gid", group)
	if err != nil {
		return nil, err
	}

	g := dsParseGroup(data)
	if g.Name != "" && g.Gid != "" {
		return g, nil
	}

	return nil, errNotFound
}
