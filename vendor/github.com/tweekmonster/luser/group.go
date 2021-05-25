// +build go1.7

package luser

import "os/user"

// Group represents a grouping of users.  Embedded *user.Group reference:
// https://golang.org/pkg/os/user/#Group
type Group struct {
	*user.Group

	// IsLuser is a flag indicating if the user was found without cgo.
	IsLuser bool
}

// LookupGroup looks up a group by name. If the group cannot be found, the
// returned error is of type UnknownGroupError.
func LookupGroup(name string) (*Group, error) {
	if fallbackEnabled {
		return lookupGroup(name)
	}

	g, err := user.LookupGroup(name)
	if err == nil {
		return &Group{Group: g}, err
	}

	return nil, err
}

// LookupGroupId looks up a group by groupid. If the group cannot be found, the
// returned error is of type UnknownGroupIdError.
func LookupGroupId(gid string) (*Group, error) {
	if fallbackEnabled {
		return lookupGroupId(gid)
	}

	g, err := user.LookupGroupId(gid)
	if err == nil {
		return &Group{Group: g}, err
	}

	return nil, err
}

// GroupIds returns the list of group IDs that the user is a member of.
func (u *User) GroupIds() ([]string, error) {
	if u.IsLuser {
		return u.lookupUserGroupIds()
	}
	return u.User.GroupIds()
}
