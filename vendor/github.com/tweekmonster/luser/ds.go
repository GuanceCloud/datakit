package luser

import (
	"bufio"
	"bytes"
	"os/exec"
	"os/user"
	"strings"
)

var dscacheutilExe = "" // path to 'dscacheutil' program.

func init() {
	if path, err := exec.LookPath("dscacheutil"); err == nil {
		dscacheutilExe = path
	}
}

func dsParseUser(data []byte) *user.User {
	u := &user.User{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.Index(line, ": "); i >= 0 {
			value := line[i+2:]

			switch line[:i] {
			case "name":
				u.Username = value
			case "uid":
				u.Uid = value
			case "gid":
				u.Gid = value
			case "dir":
				u.HomeDir = value
			case "gecos":
				if i := strings.Index(value, ","); i >= 0 {
					value = value[:i]
				}
				u.Name = value
			}
		}
	}

	return u
}

func dsUser(username string) (*user.User, error) {
	data, err := command(dscacheutilExe, "-q", "user", "-a", "name", username)
	if err != nil {
		return nil, err
	}

	u := dsParseUser(data)

	if u.Uid != "" && u.Username != "" && u.HomeDir != "" {
		return u, nil
	}

	return nil, errNotFound
}

func dsUserId(uid string) (*user.User, error) {
	data, err := command(dscacheutilExe, "-q", "user", "-a", "uid", uid)
	if err != nil {
		return nil, err
	}

	u := dsParseUser(data)

	if u.Uid != "" && u.Username != "" && u.HomeDir != "" {
		return u, nil
	}

	return nil, errNotFound
}
