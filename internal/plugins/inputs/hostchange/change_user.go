// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

const (
	FilePathPasswd  = "/etc/passwd"
	FilePathShadow  = "/etc/shadow"
	FilePathGroup   = "/etc/group"
	FilePathGShadow = "/etc/gshadow"

	// Individual change IDs.
	ChangeIDCreateUser          changes.ChangeID = "host_change_01_01"
	ChangeIDDeleteUser          changes.ChangeID = "host_change_01_02"
	ChangeIDModifyUser          changes.ChangeID = "host_change_01_03"
	ChangeIDCreateGroup         changes.ChangeID = "host_change_01_04"
	ChangeIDDeleteGroup         changes.ChangeID = "host_change_01_05"
	ChangeIDModifyGroup         changes.ChangeID = "host_change_01_06"
	ChangeIDAddUserToGroup      changes.ChangeID = "host_change_01_07"
	ChangeIDRemoveUserFromGroup changes.ChangeID = "host_change_01_08"
)

type UserChangeItem struct {
	ChangeID changes.ChangeID `toml:"change_id"` // Change ID
	UserName,
	HomeDir,
	ChangeField,
	LoginShell string
	GroupName string
	UID,
	GID int
	NewValue,
	OldValue any
	ChangeTimestampMicro int64 // Change timestamp in microseconds
}

// UserFileChanges stores user changes for a single category (user or group).
type UserFileChanges struct {
	Created  []UserChangeItem
	Deleted  []UserChangeItem
	Modified []UserChangeItem
}

// UserChangesByType stores all user changes grouped by change type.
type UserChangesByType map[string]*UserFileChanges

type UserGroup struct {
	Enabled bool `toml:"enabled"` // Enable user and group change detection

	userInfo    *Change[ChangeKey, interface{}]
	passwdPath  string
	shadowPath  string
	groupPath   string
	gShadowPath string

	// Last file modification times to avoid unnecessary parsing
	lastPasswdModTime  time.Time
	lastShadowModTime  time.Time
	lastGroupModTime   time.Time
	lastGShadowModTime time.Time

	ipt *Input
}

func (ug *UserGroup) Init(ipt *Input) error {
	if ipt == nil {
		return fmt.Errorf("input is nil, cannot init UserGroup")
	}

	ug.ipt = ipt

	ug.passwdPath = FilePathPasswd
	ug.shadowPath = FilePathShadow
	ug.groupPath = FilePathGroup
	ug.gShadowPath = FilePathGShadow

	return nil
}

func (ug *UserGroup) Collect() ([]*ChangeItem, error) {
	change, err := ug.fetchUserInfoAndEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	defer func() {
		ug.userInfo = change
	}()

	return ug.getChangeEvents(change)
}

func (ug *UserGroup) add(changesByType UserChangesByType) GetChangeItemFunc[ChangeKey, interface{}] {
	return func(key ChangeKey, newValues, oldValues []interface{}, parentChanges ...*Change[ChangeKey, interface{}]) *ChangeItem {
		// Extract the user or group name from the key
		keyStr := string(key)
		var changeType string
		var name string
		switch {
		case strings.HasPrefix(keyStr, "user:"):
			changeType = string(ChangeIDCreateUser)
			name = keyStr[5:]
		case strings.HasPrefix(keyStr, "group:"):
			changeType = string(ChangeIDCreateGroup)
			name = keyStr[6:]
		default:
			return nil
		}

		// Get or create the change group
		if changesByType[changeType] == nil {
			changesByType[changeType] = &UserFileChanges{}
		}

		// Create change items for new values
		for _, value := range newValues {
			var changeItem UserChangeItem
			if user, ok := value.(User); ok {
				changeItem = UserChangeItem{
					ChangeID:             ChangeIDCreateUser,
					UserName:             name,
					GID:                  user.GID,
					UID:                  user.UID,
					HomeDir:              user.HomeDir,
					LoginShell:           user.Shell,
					ChangeTimestampMicro: time.Now().UnixMicro(),
				}
			} else if group, ok := value.(Group); ok {
				changeItem = UserChangeItem{
					ChangeID:             ChangeIDCreateGroup,
					GroupName:            name,
					GID:                  group.GID,
					ChangeTimestampMicro: time.Now().UnixMicro(),
				}
			}
			changesByType[changeType].Created = append(changesByType[changeType].Created, changeItem)
		}

		return nil
	}
}

func (ug *UserGroup) delete(changesByType UserChangesByType) GetChangeItemFunc[ChangeKey, interface{}] {
	return func(key ChangeKey, newValues, oldValues []interface{}, parentChanges ...*Change[ChangeKey, interface{}]) *ChangeItem {
		// Extract the user or group name from the key
		keyStr := string(key)
		var changeType string
		var name string
		switch {
		case strings.HasPrefix(keyStr, "user:"):
			changeType = string(ChangeIDDeleteUser)
			name = keyStr[5:]
		case strings.HasPrefix(keyStr, "group:"):
			changeType = string(ChangeIDDeleteGroup)
			name = keyStr[6:]
		default:
			return nil
		}

		// Get or create the change group
		if changesByType[changeType] == nil {
			changesByType[changeType] = &UserFileChanges{}
		}

		// Create change items for deleted values
		for _, value := range oldValues {
			var changeItem UserChangeItem
			if user, ok := value.(User); ok {
				changeItem = UserChangeItem{
					ChangeID:             ChangeIDDeleteUser,
					UserName:             name,
					GID:                  user.GID,
					UID:                  user.UID,
					HomeDir:              user.HomeDir,
					LoginShell:           user.Shell,
					ChangeTimestampMicro: time.Now().UnixMicro(),
				}
			} else if group, ok := value.(Group); ok {
				changeItem = UserChangeItem{
					ChangeID:             ChangeIDDeleteGroup,
					GroupName:            name,
					GID:                  group.GID,
					ChangeTimestampMicro: time.Now().UnixMicro(),
				}
			}
			changesByType[changeType].Deleted = append(changesByType[changeType].Deleted, changeItem)
		}

		return nil
	}
}

func (ug *UserGroup) modify(changesByType UserChangesByType) GetChangeItemFunc[ChangeKey, interface{}] {
	return func(key ChangeKey, newValues, oldValues []interface{}, parentChanges ...*Change[ChangeKey, interface{}]) *ChangeItem {
		// Extract the user or group name from the key
		keyStr := string(key)
		var name string
		switch {
		case strings.HasPrefix(keyStr, "user:"):
			name = keyStr[5:]
		case strings.HasPrefix(keyStr, "group:"):
			name = keyStr[6:]
		default:
			return nil
		}

		// Only compare if we have exactly one old and one new value
		if len(newValues) != 1 || len(oldValues) != 1 {
			return nil
		}

		newVal := newValues[0]
		oldVal := oldValues[0]

		// Compare user changes
		if newUser, ok := newVal.(User); ok {
			if oldUser, ok := oldVal.(User); ok {
				// Compare user attributes
				if newUser.UID != oldUser.UID {
					changeItem := buildUserModifyEvent(time.Now(), name, "UID", oldUser.UID, newUser.UID)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.GID != oldUser.GID {
					changeItem := buildUserModifyEvent(time.Now(), name, "GID", oldUser.GID, newUser.GID)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.HomeDir != oldUser.HomeDir {
					changeItem := buildUserModifyEvent(time.Now(), name, "home directory", oldUser.HomeDir, newUser.HomeDir)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.Shell != oldUser.Shell {
					changeItem := buildUserModifyEvent(time.Now(), name, "login shell", oldUser.Shell, newUser.Shell)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.PasswordHash != oldUser.PasswordHash {
					changeItem := buildUserModifyEvent(time.Now(), name, "password hash", oldUser.PasswordHash, newUser.PasswordHash)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.PasswordExp != oldUser.PasswordExp {
					changeItem := buildUserModifyEvent(time.Now(), name, "password expiration time", oldUser.PasswordExp, newUser.PasswordExp)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
				if newUser.IsLocked != oldUser.IsLocked {
					changeItem := buildUserModifyEvent(time.Now(), name, "lock status", oldUser.IsLocked, newUser.IsLocked)
					changeType := string(ChangeIDModifyUser)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}
			}
		}

		// Compare group changes
		if newGroup, ok := newVal.(Group); ok {
			if oldGroup, ok := oldVal.(Group); ok {
				if newGroup.GID != oldGroup.GID {
					changeItem := buildGroupModifyEvent(time.Now(), name, "GID", oldGroup.GID, newGroup.GID)
					changeType := string(ChangeIDModifyGroup)
					if changesByType[changeType] == nil {
						changesByType[changeType] = &UserFileChanges{}
					}
					changesByType[changeType].Modified = append(changesByType[changeType].Modified, changeItem)
				}

				// Compare group members
				oldMembersMap := make(map[string]bool)
				for _, m := range oldGroup.Members {
					oldMembersMap[m] = true
				}
				for _, m := range newGroup.Members {
					if !oldMembersMap[m] {
						changeItem := UserChangeItem{
							ChangeID:             ChangeIDAddUserToGroup,
							GroupName:            name,
							GID:                  newGroup.GID,
							UserName:             m,
							ChangeTimestampMicro: time.Now().UnixMicro(),
						}
						changeType := string(ChangeIDAddUserToGroup)
						if changesByType[changeType] == nil {
							changesByType[changeType] = &UserFileChanges{}
						}
						changesByType[changeType].Created = append(changesByType[changeType].Created, changeItem)
					}
				}

				currentMembersMap := make(map[string]bool)
				for _, m := range newGroup.Members {
					currentMembersMap[m] = true
				}
				for _, m := range oldGroup.Members {
					if !currentMembersMap[m] {
						changeItem := UserChangeItem{
							ChangeID:             ChangeIDRemoveUserFromGroup,
							GroupName:            name,
							GID:                  newGroup.GID,
							UserName:             m,
							ChangeTimestampMicro: time.Now().UnixMicro(),
						}
						changeType := string(ChangeIDRemoveUserFromGroup)
						if changesByType[changeType] == nil {
							changesByType[changeType] = &UserFileChanges{}
						}
						changesByType[changeType].Deleted = append(changesByType[changeType].Deleted, changeItem)
					}
				}
			}
		}

		return nil
	}
}

func (ug *UserGroup) getChangeItems(userChangeItems []UserChangeItem) ([]*ChangeItem, error) {
	// Group changes by their type
	changesByType := make(map[string][]UserChangeItem)
	for _, change := range userChangeItems {
		changesByType[string(change.ChangeID)] = append(changesByType[string(change.ChangeID)], change)
	}

	var changeItems []*ChangeItem

	// Process each change type group
	for changeType, changeList := range changesByType {
		// Use original change ID
		originalChangeID := changes.ChangeID(changeType)
		var changeTimestampMicro int64
		if len(changeList) > 0 {
			changeTimestampMicro = changeList[0].ChangeTimestampMicro
		}

		// Create merged data for template rendering
		mergedData := struct {
			ChangeCount int              `toml:"change_count"` // Number of changes
			Changes     []UserChangeItem `toml:"changes"`      // List of individual changes
		}{
			ChangeCount: len(changeList),
			Changes:     changeList,
		}

		// Render template with merged data
		title, message, err := changes.RenderHostTemplate(defaultChangeLanguage, originalChangeID, mergedData)
		if err != nil {
			l.Warnf("RenderHostTemplate failed for %s: %v, using default message", string(originalChangeID), err.Error())
			// Fallback to default title and message
			title = fmt.Sprintf("User/Group Changes (%s)", changeType)
			message = fmt.Sprintf("Detected %d user/group change(s)", len(changeList))
		}

		// Add to result
		changeItems = append(changeItems, &ChangeItem{
			ChangeID:             originalChangeID,
			ChangeTimestampMicro: changeTimestampMicro,
			Title:                title,
			Message:              message,
		})
	}

	return changeItems, nil
}

func (ug *UserGroup) getChangeEvents(change *Change[ChangeKey, interface{}]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}
	changesByType := UserChangesByType{}

	if ug.userInfo != nil {
		// Collect changes using GetChangeEvent
		_, err := change.GetChangeEvent(ug.userInfo, &ChangeItemConfig[ChangeKey, interface{}]{
			Add:    ug.add(changesByType),
			Delete: ug.delete(changesByType),
			Modify: ug.modify(changesByType),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get user change event: %w", err)
		}
	}

	// Convert collected changes to ChangeItem
	var allUserChangeItems []UserChangeItem
	for _, changes := range changesByType {
		allUserChangeItems = append(allUserChangeItems, changes.Created...)
		allUserChangeItems = append(allUserChangeItems, changes.Deleted...)
		allUserChangeItems = append(allUserChangeItems, changes.Modified...)
	}

	// Process all changes using existing getChangeItems method
	if len(allUserChangeItems) > 0 {
		processedChanges, err := ug.getChangeItems(allUserChangeItems)
		if err != nil {
			return nil, fmt.Errorf("failed to process user change items: %w", err)
		}
		changeEvents = append(changeEvents, processedChanges...)
	}

	return changeEvents, nil
}

// Config configuration structure.
type Config struct {
	CheckInterval  time.Duration // Check interval
	CachePath      string        // Baseline cache path
	HighRiskUsers  []string      // High-risk users list
	HighRiskGroups []string      // High-risk groups list
}

// User user structure (parsed from /etc/passwd + /etc/shadow).
type User struct {
	UserName     string `json:"user_name"`     // Username
	UID          int    `json:"uid"`           // User ID
	GID          int    `json:"gid"`           // Primary group ID
	Comment      string `json:"comment"`       // Comment information
	HomeDir      string `json:"home_dir"`      // Home directory
	Shell        string `json:"shell"`         // Login shell
	PasswordHash string `json:"password_hash"` // Password hash (from /etc/shadow)
	PasswordExp  string `json:"password_exp"`  // Password expiration time (from /etc/shadow)
	IsLocked     bool   `json:"is_locked"`     // Whether locked (from /etc/shadow)
}

// Group group structure (parsed from /etc/group + /etc/gshadow).
type Group struct {
	GroupName string   `json:"group_name"` // Group name
	GID       int      `json:"gid"`        // Group ID
	Members   []string `json:"members"`    // Member list (secondary groups)
	Password  string   `json:"password"`   // Group password (from /etc/gshadow)
	Admins    []string `json:"admins"`     // Group administrators (from /etc/gshadow)
}

func (ug *UserGroup) fetchUserInfoAndEvents() (*Change[ChangeKey, interface{}], error) {
	isShadowFileExist := true
	isGroupShadowFileExist := true

	// Get current file modification times
	passwdModTime, err := GetFileModTime(ug.passwdPath)
	if err != nil {
		return nil, err
	}

	shadowModTime, err := GetFileModTime(ug.shadowPath)
	if os.IsNotExist(err) {
		isShadowFileExist = false
	} else if err != nil {
		return nil, err
	}

	gShadowModTime, err := GetFileModTime(ug.gShadowPath)
	if os.IsNotExist(err) {
		isGroupShadowFileExist = false
	} else if err != nil {
		return nil, err
	}

	groupModTime, err := GetFileModTime(ug.groupPath)
	if err != nil {
		return nil, err
	}

	// Create a new Change instance for user info
	newUserInfo := NewChange[ChangeKey, interface{}]()
	if !passwdModTime.Equal(ug.lastPasswdModTime) || (isShadowFileExist && !shadowModTime.Equal(ug.lastShadowModTime)) {
		// Parse user info
		users, err := ParsePasswd(ug.passwdPath)
		if err != nil {
			return nil, err
		}
		if isShadowFileExist {
			if err := ParseShadow(users, ug.shadowPath); err != nil {
				return nil, err
			}
		}

		// Add users to Change instance
		for userName, user := range users {
			newUserInfo.Add(ChangeKey("user:"+userName), user)
		}

		// Update last modification times
		ug.lastPasswdModTime = passwdModTime
		if isShadowFileExist {
			ug.lastShadowModTime = shadowModTime
		}
	} else if ug.userInfo != nil {
		// Copy existing user info if no changes
		for key, values := range ug.userInfo.data {
			if strings.HasPrefix(string(key), "user:") {
				for _, value := range values {
					newUserInfo.Add(key, value)
				}
			}
		}
	}
	if !groupModTime.Equal(ug.lastGroupModTime) || (isGroupShadowFileExist && !gShadowModTime.Equal(ug.lastGShadowModTime)) {
		// Parse group info
		groups, err := ParseGroup(ug.groupPath)
		if err != nil {
			return nil, err
		}
		if isGroupShadowFileExist {
			if err := ParseGShadow(groups, ug.gShadowPath); err != nil {
				return nil, err
			}
		}

		// Add groups to Change instance
		for groupName, group := range groups {
			newUserInfo.Add(ChangeKey("group:"+groupName), group)
		}

		// Update last modification times
		ug.lastGroupModTime = groupModTime
		if isGroupShadowFileExist {
			ug.lastGShadowModTime = gShadowModTime
		}
	} else if ug.userInfo != nil {
		// Copy existing group info if no changes
		for key, values := range ug.userInfo.data {
			if strings.HasPrefix(string(key), "group:") {
				for _, value := range values {
					newUserInfo.Add(key, value)
				}
			}
		}
	}

	return newUserInfo, nil
}

// Build user modification event.
func buildUserModifyEvent(changeTime time.Time, userName, field string, oldVal, newVal interface{}) UserChangeItem {
	return UserChangeItem{
		ChangeID:             ChangeIDModifyUser,
		UserName:             userName,
		ChangeField:          field,
		OldValue:             oldVal,
		NewValue:             newVal,
		ChangeTimestampMicro: changeTime.UnixMicro(),
	}
}

// Build group modification event.
func buildGroupModifyEvent(changeTime time.Time, groupName, field string, oldVal, newVal interface{}) UserChangeItem {
	return UserChangeItem{
		GroupName:            groupName,
		ChangeID:             ChangeIDModifyGroup,
		ChangeField:          field,
		OldValue:             oldVal,
		NewValue:             newVal,
		ChangeTimestampMicro: changeTime.UnixMicro(),
	}
}

// ParsePasswd parse passwd file.
func ParsePasswd(passwdPath string) (map[string]User, error) {
	content, err := ReadFile(passwdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", passwdPath, err)
	}

	users := make(map[string]User)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 7 {
			continue // invalid format, skip
		}

		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])

		user := User{
			UserName: parts[0],
			UID:      uid,
			GID:      gid,
			Comment:  parts[4],
			HomeDir:  parts[5],
			Shell:    parts[6],
		}
		users[parts[0]] = user
	}

	return users, nil
}

// ParseShadow parse shadow file (supplement user password information).
func ParseShadow(users map[string]User, shadowPath string) error {
	content, err := ReadFile(shadowPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", shadowPath, err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 9 {
			continue
		}

		userName := parts[0]
		if user, exists := users[userName]; exists {
			user.PasswordHash = parts[1]
			user.PasswordExp = parts[2]
			user.IsLocked = parts[1] == "!" || parts[1] == "*" // lock indicator
			users[userName] = user
		}
	}

	return nil
}

// ParseGroup parse group file.
func ParseGroup(groupPath string) (map[string]Group, error) {
	content, err := ReadFile(groupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", groupPath, err)
	}

	groups := make(map[string]Group)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			continue
		}

		gid, _ := strconv.Atoi(parts[2])
		members := []string{}
		if parts[3] != "" {
			members = strings.Split(parts[3], ",")
		}

		group := Group{
			GroupName: parts[0],
			GID:       gid,
			Members:   members,
		}
		groups[parts[0]] = group
	}

	return groups, nil
}

// ParseGShadow parse gshadow file (supplement group password information).
func ParseGShadow(groups map[string]Group, gShadowPath string) error {
	content, err := ReadFile(gShadowPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", gShadowPath, err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			continue
		}

		groupName := parts[0]
		if group, exists := groups[groupName]; exists {
			group.Password = parts[1]
			admins := []string{}
			if parts[3] != "" {
				admins = strings.Split(parts[3], ",")
			}
			group.Admins = admins
			groups[groupName] = group
		}
	}

	return nil
}
