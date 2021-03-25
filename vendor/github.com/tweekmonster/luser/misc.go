package luser

import (
	"errors"
	"os/exec"
	"os/user"
)

// ErrListGroups returned when LookupGroupId() has no fallback.
var ErrListGroups = errors.New("luser: unable to list groups")

// ErrCurrentUser returned when Current() fails to get the user.
var ErrCurrentUser = errors.New("luser: unable to get current user")

// Generic internal error indicating that a search function did its job but
// still couldn't find a user/group.
var errNotFound = errors.New("not found")

// Wrap user.User with luser.User
func luser(u *user.User) *User {
	return &User{User: u, IsLuser: true}
}

// Convenience function for running a program and returning the output.
func command(program string, args ...string) ([]byte, error) {
	cmd := exec.Command(program, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
