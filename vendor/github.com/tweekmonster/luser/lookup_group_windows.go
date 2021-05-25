package luser

func lookupGroup(name string) (*Group, error) {
	return nil, errWin
}

func lookupGroupId(gid string) (*Group, error) {
	return nil, errWin
}

func (u *User) lookupUserGroupIds() ([]string, error) {
	return nil, ErrListGroups
}
