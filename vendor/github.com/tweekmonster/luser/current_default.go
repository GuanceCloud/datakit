// +build !windows
// +build !cgo

package luser

import (
	"os"
	"os/user"
	"strconv"
	"sync"
)

var currentUid = -1
var current *User
var currentMu sync.Mutex

func init() {
	fallbackEnabled = true
}

// The default stub uses $USER and $HOME to fill in the user.  A more reliable
// method will be tried before falling back to the stub.
func currentUser() (*User, error) {
	uid := os.Getuid()
	if uid >= 0 {
		if uid == currentUid && current != nil {
			return current, nil
		}

		currentMu.Lock()
		defer currentMu.Unlock()

		currentUid = uid
		u, err := lookupId(strconv.Itoa(uid))
		current = u
		if err == nil {
			return u, nil
		}
	}

	if u, err := user.Current(); err == nil {
		return luser(u), nil
	}

	return nil, ErrCurrentUser
}
