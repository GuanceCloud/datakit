// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUserGroup_Init tests UserGroup initialization
func TestUserGroup_Init(t *testing.T) {
	ug := &UserGroup{
		Enabled: true,
	}

	ipt := &Input{}

	// Test initialization
	err := ug.Init(ipt)
	assert.NoError(t, err)

	// Verify initialization
	assert.Equal(t, ipt, ug.ipt)
	assert.Equal(t, FilePathPasswd, ug.passwdPath)
	assert.Equal(t, FilePathShadow, ug.shadowPath)
	assert.Equal(t, FilePathGroup, ug.groupPath)
	assert.Equal(t, FilePathGShadow, ug.gShadowPath)

	// Test initialization with nil input
	err = ug.Init(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input is nil")
}

// TestUserGroup_CompareUsers tests user comparison functionality
func TestUserGroup_CompareUsers(t *testing.T) {
	// Create a UserGroup instance
	ug := &UserGroup{}

	// Init the UserGroup to set default file paths
	err := ug.Init(&Input{})
	assert.NoError(t, err)

	// Mock old user data
	oldUsers := map[string]User{
		"user1": {
			UserName: "user1",
			UID:      1001,
			GID:      1001,
			HomeDir:  "/home/user1",
			Shell:    "/bin/bash",
		},
		"user2": {
			UserName: "user2",
			UID:      1002,
			GID:      1002,
			HomeDir:  "/home/user2",
			Shell:    "/bin/bash",
		},
		"ignoreduser": {
			UserName: "ignoreduser",
			UID:      1003,
			GID:      1003,
			HomeDir:  "/home/ignoreduser",
			Shell:    "/bin/bash",
		},
	}

	// Mock new user data with some changes
	newUsers := map[string]User{
		"user1": {
			UserName: "user1",
			UID:      1001,
			GID:      1001,
			HomeDir:  "/home/user1_updated", // Changed home directory
			Shell:    "/bin/bash",
		},
		"user3": {
			UserName: "user3",
			UID:      1003,
			GID:      1003,
			HomeDir:  "/home/user3", // New user
			Shell:    "/bin/bash",
		},
		"ignoreduser": {
			UserName: "ignoreduser",
			UID:      1004,
			GID:      1004,
			HomeDir:  "/home/ignoreduser_updated", // Changed, but should be ignored
			Shell:    "/bin/bash",
		},
	}

	// Create Change instances for comparison
	oldChange := NewChange[ChangeKey, interface{}]()
	for userName, user := range oldUsers {
		oldChange.Add(ChangeKey("user:"+userName), user)
	}

	newChange := NewChange[ChangeKey, interface{}]()
	for userName, user := range newUsers {
		newChange.Add(ChangeKey("user:"+userName), user)
	}

	// Set the old user info
	ug.userInfo = oldChange

	changeItems, err := ug.getChangeEvents(newChange)
	assert.NoError(t, err)

	// Verify results
	assert.NotEmpty(t, changeItems)
	// We should have at least one change for user1's home directory change
	// and one for user3's creation, user2's deletion
	assert.GreaterOrEqual(t, len(changeItems), 1)
}

// TestUserGroup_CompareGroups tests group comparison functionality
func TestUserGroup_CompareGroups(t *testing.T) {
	// Create a UserGroup instance
	ug := &UserGroup{}

	// Init the UserGroup to set default file paths
	err := ug.Init(&Input{})
	assert.NoError(t, err)

	// Create old and new group maps
	oldGroups := map[string]Group{
		"group1": {
			GroupName: "group1",
			GID:       2001,
			Members:   []string{"user1", "user2"},
		},
		"group2": {
			GroupName: "group2",
			GID:       2002,
			Members:   []string{"user1"},
		},
		"ignoredgroup": {
			GroupName: "ignoredgroup",
			GID:       2003,
			Members:   []string{"user1"},
		},
	}

	newGroups := map[string]Group{
		"group1": {
			GroupName: "group1",
			GID:       2001,
			Members:   []string{"user1", "user3"}, // Changed members
		},
		"group3": {
			GroupName: "group3",
			GID:       2003,
			Members:   []string{"user1", "user2"},
		},
		"ignoredgroup": {
			GroupName: "ignoredgroup",
			GID:       2003,
			Members:   []string{"user1", "ignoreduser"}, // Ignored user added
		},
	}

	// Create Change instances for comparison
	oldChange := NewChange[ChangeKey, interface{}]()
	for groupName, group := range oldGroups {
		oldChange.Add(ChangeKey("group:"+groupName), group)
	}

	newChange := NewChange[ChangeKey, interface{}]()
	for groupName, group := range newGroups {
		newChange.Add(ChangeKey("group:"+groupName), group)
	}

	// Set the old user info
	ug.userInfo = oldChange

	changeItems, err := ug.getChangeEvents(newChange)
	assert.NoError(t, err)

	// Verify results
	assert.NotEmpty(t, changeItems)
	assert.GreaterOrEqual(t, len(changeItems), 1)
}

// TestBuildUserModifyEvent tests building user modification event
func TestBuildUserModifyEvent(t *testing.T) {
	changeTime := time.Now()
	event := buildUserModifyEvent(changeTime, "testuser", "UID", 1000, 1001)

	assert.Equal(t, ChangeIDModifyUser, event.ChangeID)
	assert.Equal(t, "testuser", event.UserName)
	assert.Equal(t, "UID", event.ChangeField)
	assert.Equal(t, 1000, event.OldValue)
	assert.Equal(t, 1001, event.NewValue)
	assert.Equal(t, changeTime.UnixMicro(), event.ChangeTimestampMicro)
}

// TestBuildGroupModifyEvent tests building group modification event
func TestBuildGroupModifyEvent(t *testing.T) {
	changeTime := time.Now()
	event := buildGroupModifyEvent(changeTime, "testgroup", "GID", 2000, 2001)

	assert.Equal(t, ChangeIDModifyGroup, event.ChangeID)
	assert.Equal(t, "testgroup", event.GroupName)
	assert.Equal(t, "GID", event.ChangeField)
	assert.Equal(t, 2000, event.OldValue)
	assert.Equal(t, 2001, event.NewValue)
	assert.Equal(t, changeTime.UnixMicro(), event.ChangeTimestampMicro)
}

// TestMaxTime tests MaxTime function
func TestMaxTime(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(1 * time.Hour)
	time3 := time1.Add(-1 * time.Hour)

	assert.Equal(t, time2, MaxTime(time1, time2))
	assert.Equal(t, time1, MaxTime(time1, time3))
	assert.Equal(t, time1, MaxTime(time1, time1))
}

// TestParsePasswd tests parsing passwd file
func TestParsePasswd(t *testing.T) {
	// Create a temporary passwd file
	passwdContent := "testuser:x:1000:1000:Test User:/home/testuser:/bin/bash\ntestuser2:x:1001:1001:Test User 2:/home/testuser2:/bin/sh"
	tmpFile, err := os.CreateTemp("", "passwd")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(passwdContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Parse passwd file
	users, err := ParsePasswd(tmpFile.Name())
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	// Check first user
	user1, exists := users["testuser"]
	assert.True(t, exists)
	assert.Equal(t, "testuser", user1.UserName)
	assert.Equal(t, 1000, user1.UID)
	assert.Equal(t, 1000, user1.GID)
	assert.Equal(t, "Test User", user1.Comment)
	assert.Equal(t, "/home/testuser", user1.HomeDir)
	assert.Equal(t, "/bin/bash", user1.Shell)

	// Check second user
	user2, exists := users["testuser2"]
	assert.True(t, exists)
	assert.Equal(t, "testuser2", user2.UserName)
	assert.Equal(t, 1001, user2.UID)
	assert.Equal(t, 1001, user2.GID)
	assert.Equal(t, "Test User 2", user2.Comment)
	assert.Equal(t, "/home/testuser2", user2.HomeDir)
	assert.Equal(t, "/bin/sh", user2.Shell)
}

// TestParseGroup tests parsing group file
func TestParseGroup(t *testing.T) {
	// Create a temporary group file
	groupContent := "testgroup:x:1000:testuser\ntestgroup2:x:1001:testuser2,testuser"
	tmpFile, err := os.CreateTemp("", "group")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(groupContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Parse group file
	groups, err := ParseGroup(tmpFile.Name())
	assert.NoError(t, err)
	assert.Len(t, groups, 2)

	// Check first group
	group1, exists := groups["testgroup"]
	assert.True(t, exists)
	assert.Equal(t, "testgroup", group1.GroupName)
	assert.Equal(t, 1000, group1.GID)
	assert.Equal(t, []string{"testuser"}, group1.Members)

	// Check second group
	group2, exists := groups["testgroup2"]
	assert.True(t, exists)
	assert.Equal(t, "testgroup2", group2.GroupName)
	assert.Equal(t, 1001, group2.GID)
	assert.Equal(t, []string{"testuser2", "testuser"}, group2.Members)
}

func TestUserGroup_Collect_mergedEvents(t *testing.T) {
	// Create a UserGroup instance
	ug := &UserGroup{
		userInfo: NewChange[ChangeKey, interface{}](), // Initialize with empty Change instance
	}

	// Init the UserGroup to set default file paths
	err := ug.Init(&Input{})
	assert.NoError(t, err)

	// Mock user info with some initial users and groups
	oldUserInfo := NewChange[ChangeKey, interface{}]()
	oldUserInfo.Add(ChangeKey("user:existinguser"), User{UserName: "existinguser", UID: 1000, GID: 1000, HomeDir: "/home/existinguser", Shell: "/bin/bash"})
	oldUserInfo.Add(ChangeKey("group:existinggroup"), Group{GroupName: "existinggroup", GID: 2000, Members: []string{"existinguser"}})

	// Create new user info with changes
	newUserInfo := NewChange[ChangeKey, interface{}]()
	newUserInfo.Add(ChangeKey("user:existinguser"), User{UserName: "existinguser", UID: 1000, GID: 1000, HomeDir: "/home/existinguser", Shell: "/bin/sh"}) // Shell changed
	newUserInfo.Add(ChangeKey("user:newuser1"), User{UserName: "newuser1", UID: 1001, GID: 1001, HomeDir: "/home/newuser1", Shell: "/bin/bash"})           // New user
	newUserInfo.Add(ChangeKey("user:newuser2"), User{UserName: "newuser2", UID: 1002, GID: 1002, HomeDir: "/home/newuser2", Shell: "/bin/bash"})           // New user
	newUserInfo.Add(ChangeKey("group:existinggroup"), Group{GroupName: "existinggroup", GID: 2000, Members: []string{"existinguser", "newuser1"}})         // Added user
	newUserInfo.Add(ChangeKey("group:newgroup"), Group{GroupName: "newgroup", GID: 2001, Members: []string{"existinguser"}})                               // New group

	// Set up user info so we can test getChangeEvents directly
	ug.userInfo = oldUserInfo

	changeItems, err := ug.getChangeEvents(newUserInfo)
	assert.NoError(t, err)
	assert.NotEmpty(t, changeItems)
}
