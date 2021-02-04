package luser

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// GetentParseFiles tells the getent fallback to prefer parsing the
// '/etc/passwd' and '/etc/group' files instead of executing the 'getent'
// program.  This is on by default since the speed of parsing the files
// directly is comparible to the C API and is many times faster than executing
// the 'getent' program.  If false, the files will still be parsed if 'getent'
// fails or isn't found.
var GetentParseFiles = true

var (
	getentExe      = "" // path to the 'getent' program.
	passwdFilePath = "" // path to the 'passwd' file.
	groupFilePath  = "" // path to the 'group' file.
)

func init() {
	if path, err := exec.LookPath("getent"); err == nil {
		getentExe = path
	}

	if _, err := os.Stat("/etc/passwd"); err == nil {
		passwdFilePath = "/etc/passwd"
	}

	if _, err := os.Stat("/etc/group"); err == nil {
		groupFilePath = "/etc/group"
	}
}

// Lookup user by parsing an database file first.  If that fails, try with the
// 'getent' program.
func getent(database, key string) (string, error) {
	if !GetentParseFiles && getentExe != "" {
		data, err := command(getentExe, database, key)
		if err == nil {
			return string(bytes.TrimSpace(data)), nil
		}
	}

	dbfile := ""
	switch database {
	case "passwd":
		dbfile = passwdFilePath
	case "group":
		dbfile = groupFilePath
	}

	if dbfile != "" {
		if line, err := searchEntityDatabase(dbfile, key); err == nil {
			return line, nil
		}
	}

	return "", errNotFound
}

func getentUser(key string) (*user.User, error) {
	entity, err := getent("passwd", key)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(entity, ":")
	if len(parts) < 6 {
		return nil, errNotFound
	}

	// Get name from GECOS.
	name := parts[4]
	if i := strings.Index(name, ","); i >= 0 {
		name = name[:i]
	}

	return &user.User{
		Username: parts[0],
		Uid:      parts[2],
		Gid:      parts[3],
		Name:     name,
		HomeDir:  parts[5],
	}, nil
}

func entityIndex(line string, index int) int {
	f := func(c rune) bool {
		if c == ':' {
			index--
			return index == 0
		}
		return false
	}

	return strings.IndexFunc(line, f)
}

// Searches an entity database for a matching name or id.
func searchEntityDatabase(filename, match string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	id := false
	if _, err := strconv.Atoi(match); err == nil {
		id = true
	}

	prefix := match + ":"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if id {
			i := entityIndex(line, 2)
			if strings.HasPrefix(line[i+1:], prefix) {
				return line, nil
			}
		} else if strings.HasPrefix(line, prefix) {
			return line, nil
		}
	}

	return "", errNotFound
}
