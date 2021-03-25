package luser

import (
	"bytes"
	"os/exec"
	"strings"
)

var idExe = "" // path to the 'id' program.

func init() {
	if path, err := exec.LookPath("id"); err == nil {
		idExe = path
	}
}

func idGroupList(u *User) ([]string, error) {
	data, err := command(idExe, "-G", u.Username)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	if len(data) > 0 {
		return strings.Fields(string(data)), nil
	}

	return nil, errNotFound
}
