// Package luser is a drop-in replacement for 'os/user' which allows you to
// lookup users and groups in cross-compiled builds without 'cgo'.
//
// 'os/user' requires 'cgo' to lookup users using the target OS's API.  This is
// the most reliable way to look up user and group information.  However,
// cross-compiling means that 'os/user' will only work for the OS you're using.
// 'user.Current()' is usable when building without 'cgo', but doesn't always
// work.  The '$USER' and '$HOME' variables could be different from what you
// expect or not even exist.
//
// If you want to cross-compile a relatively simple program that needs to write
// a config file somewhere in the user's directory, the last thing you want to
// do is figure out some elaborate build scheme involving virtual machines.
//
// When cgo is not available for a build, one of the following methods will be
// used to lookup user and group information:
//
//  | Method        | Used for                                                       |
//  |---------------|----------------------------------------------------------------|
//  | `/etc/passwd` | Parsed to lookup user information. (Unix, Linux)               |
//  | `/etc/group`  | Parsed to lookup group information. (Unix, Linux)              |
//  | `getent`      | Optional. Find user/group information. (Unix, Linux)           |
//  | `dscacheutil` | Lookup user/group information via Directory Services. (Darwin) |
//  | `id`          | Finding a user's groups when using `GroupIds()`.               |
//
// You should be able to simply replace 'user.' with 'luser.' (in most cases).
package luser

import "os/user"

// Switched to true in current_default.go.  Windows does not use a fallback.
var fallbackEnabled = false

// User represents a user account.  Embedded *user.User reference:
// https://golang.org/pkg/os/user/#User
type User struct {
	*user.User

	IsLuser bool // flag indicating if the user was found without cgo.
}

// Current returns the current user.  On builds where cgo is available, this
// returns the result from user.Current().  Otherwise, alternate lookup methods
// are used before falling back to the built-in stub.
func Current() (*User, error) {
	return currentUser()
}

// Lookup looks up a user by username. If the user cannot be found, the
// returned error is of type UnknownUserError.
func Lookup(username string) (*User, error) {
	if fallbackEnabled {
		return lookupUser(username)
	}

	u, err := user.Lookup(username)
	if err == nil {
		return &User{User: u}, nil
	}

	return nil, err
}

// LookupId looks up a user by userid. If the user cannot be found, the
// returned error is of type UnknownUserIdError.
func LookupId(uid string) (*User, error) {
	if fallbackEnabled {
		return lookupId(uid)
	}

	u, err := user.LookupId(uid)
	if err == nil {
		return &User{User: u}, nil
	}

	return nil, err
}
