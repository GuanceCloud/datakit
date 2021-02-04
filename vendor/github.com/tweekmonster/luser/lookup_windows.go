// Windows stub since syscall is used instead of cgo.  os/user will work fine
// when cross-compiling Windows.  Aside from currentUser(), the other functions
// should never excute since the original errors would need to include
// "requires cgo" in the message before calling these.

package luser

import (
	"errors"
	"os/user"
)

var errWin = errors.New("luser: you should not get this error")

func currentUser() (*User, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	return luser(u), nil
}

func lookupUser(username string) (*User, error) {
	return nil, errWin
}

func lookupId(uid string) (*User, error) {
	return nil, errWin
}
