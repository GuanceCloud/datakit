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
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

// Default crontab paths to monitor.
var (
	FilePathCrontab      = "/etc/crontab"
	DirPathCronD         = "/etc/cron.d"
	DirPathSpoolCrontabs = "/var/spool/cron/crontabs"
)

// Crontab related change ID.
const (
	ChangeIDCrontab = "host_change_02_01"
)

// Crontab structure (parsed from /etc/crontab).
type Crontab struct {
	User     string `json:"user"`      // User for which the crontab entry is scheduled
	Schedule string `json:"schedule"`  // Crontab schedule expression
	Command  string `json:"command"`   // Command to be executed
	FilePath string `json:"file_path"` // Path to the crontab file containing this entry
}

// CrontabChecker handles crontab change detection.
type CrontabChecker struct {
	Enabled bool `toml:"enabled"` // Enable crontab change detection

	crontabChange *Change[uint64, *Crontab]
	input         *Input
}

// CrontabInfo stores crontab information and state.
type CrontabInfo struct {
	CreateTimestampSeconds int64            `json:"create_timestamp_seconds"` // When this baseline was created
	Crontabs               []Crontab        `json:"crontabs"`                 // Crontab entries
	CrontabFileModTimes    map[string]int64 `json:"crontab_file_mod_times"`   // File modification times for all monitored crontab files
}
type CrontabFileInfo struct {
	ModTimestampSeconds int64      `json:"mod_timestamp_seconds"` // Last modification timestamp of the crontab file
	Crontabs            []*Crontab `json:"crontabs"`              // Crontab entries
	FilePath            string     `json:"file_path"`             // Path to the crontab file
	ContentHash         uint64     `json:"content_hash"`          // Hash value of the crontab content
	Content             string     `json:"content"`               // Content of the crontab file
}

// getCrontabFiles returns all crontab files to monitor.
func getCrontabFiles() ([]string, error) {
	files := []string{FilePathCrontab}

	// Add files from /etc/cron.d/
	cronDFiles, err := filepath.Glob(filepath.Join(DirPathCronD, "*"))
	if err != nil {
		l.Warnf("failed to glob %s: %w", DirPathCronD, err)
	} else {
		files = append(files, cronDFiles...)
	}

	// Add files from /var/spool/cron/crontabs/
	spoolFiles, err := filepath.Glob(filepath.Join(DirPathSpoolCrontabs, "*"))
	if err != nil {
		l.Warnf("failed to glob %s: %w", DirPathSpoolCrontabs, err)
	} else {
		files = append(files, spoolFiles...)
	}

	return files, nil
}

// CrontabChangeItem represents a crontab change event.
type CrontabChangeItem struct {
	ChangeID             changes.ChangeID `json:"change_id"`              // Unique change identifier
	ChangeTimestampMicro int64            `json:"change_timestamp_micro"` // Change timestamp in microseconds
	Title                string           `json:"title"`                  // Event title
	Message              string           `json:"message"`                // Detailed message
	CrontabUser          string           `json:"crontab_user"`           // User for which the crontab entry is scheduled
	CrontabSchedule      string           `json:"crontab_schedule"`       // Crontab schedule expression
	CrontabCommand       string           `json:"crontab_command"`        // Command to be executed
	CrontabFilePath      string           `json:"crontab_file_path"`      // Path to the crontab file

	// Fields for template rendering
	User     string `json:"user"`      // User for which the crontab entry is scheduled
	Schedule string `json:"schedule"`  // Crontab schedule expression
	Command  string `json:"command"`   // Command to be executed
	FilePath string `json:"file_path"` // Path to the crontab file
}

// CronFileChanges stores crontab changes for a single file.
type CronFileChanges struct {
	Created []*Crontab
	Deleted []*Crontab
}

// CronChangesByFile stores all crontab changes grouped by file path.
type CronChangesByFile map[string]*CronFileChanges

// CrontabMergedChangeItem represents a merged crontab change event.
type CrontabMergedChangeItem struct {
	ChangeID             changes.ChangeID  `json:"change_id"`              // Unique change identifier
	ChangeTimestampMicro int64             `json:"change_timestamp_micro"` // Change timestamp in microseconds
	Title                string            `json:"title"`                  // Event title
	Message              string            `json:"message"`                // Detailed message
	ChangesByFile        CronChangesByFile `json:"changes_by_file"`        // Changes grouped by file path
}

// Init initializes the CrontabChecker.
func (cc *CrontabChecker) Init(input *Input) error {
	cc.input = input
	return nil
}

// Collect collects crontab changes and returns them as ChangeItem.
func (cc *CrontabChecker) Collect() ([]*ChangeItem, error) {
	change, err := cc.getCrontabChange()
	if err != nil {
		return nil, fmt.Errorf("failed to get crontab change: %w", err)
	}

	defer func() {
		cc.crontabChange = change
	}()

	changeEvents, err := cc.getChangeEvents(change)
	if err != nil {
		return nil, fmt.Errorf("failed to get change events: %w", err)
	}

	return changeEvents, nil
}

func (cc *CrontabChecker) add(changesByFile map[string]*CronFileChanges) GetChangeItemFunc[uint64, *Crontab] {
	return func(key uint64, newValues, oldValues []*Crontab, parentChanges ...*Change[uint64, *Crontab]) *ChangeItem {
		var parentChange *Change[uint64, *Crontab]
		if len(parentChanges) > 0 {
			parentChange = parentChanges[0]
		}

		file := parentChange.RawKey
		if changesByFile[file] == nil {
			changesByFile[file] = &CronFileChanges{}
		}
		changesByFile[file].Created = append(changesByFile[file].Created, newValues...)
		return nil
	}
}

func (cc *CrontabChecker) delete(changesByFile map[string]*CronFileChanges) GetChangeItemFunc[uint64, *Crontab] {
	return func(key uint64, newValues, oldValues []*Crontab, parentChanges ...*Change[uint64, *Crontab]) *ChangeItem {
		var parentChange *Change[uint64, *Crontab]
		if len(parentChanges) > 0 {
			parentChange = parentChanges[0]
		}

		file := parentChange.RawKey
		if changesByFile[file] == nil {
			changesByFile[file] = &CronFileChanges{}
		}
		changesByFile[file].Deleted = append(changesByFile[file].Deleted, oldValues...)
		return nil
	}
}

func (cc *CrontabChecker) modify(changesByFile map[string]*CronFileChanges) GetChangeItemFunc[uint64, *Crontab] {
	return func(key uint64, newValues, oldValues []*Crontab, parentChanges ...*Change[uint64, *Crontab]) *ChangeItem {
		var parentChange *Change[uint64, *Crontab]
		if len(parentChanges) > 0 {
			parentChange = parentChanges[0]
		}

		file := parentChange.RawKey
		if changesByFile[file] == nil {
			changesByFile[file] = &CronFileChanges{}
		}

		newLen := len(newValues)
		oldLen := len(oldValues)
		if newLen > oldLen {
			changesByFile[file].Created = append(changesByFile[file].Created, newValues[oldLen:]...)
		} else if newLen < oldLen {
			changesByFile[file].Deleted = append(changesByFile[file].Deleted, oldValues[newLen:]...)
		}

		return nil
	}
}

func (cc *CrontabChecker) getChangeEvents(change *Change[uint64, *Crontab]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}
	changesByFile := map[string]*CronFileChanges{}

	if cc.crontabChange != nil {
		// collect changes
		_, err := change.GetChangeEvent(cc.crontabChange, &ChangeItemConfig[uint64, *Crontab]{
			Add:    cc.add(changesByFile),
			Delete: cc.delete(changesByFile),
			Modify: cc.modify(changesByFile),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get crontab change event: %w", err)
		}
	}

	//
	for file, changes := range changesByFile {
		if file == "" || (len(changes.Created) == 0 && len(changes.Deleted) == 0) {
			continue
		}
		crontabChangeEvent := cc.getChangeEvent(file, changes)
		changeEvents = append(changeEvents, &ChangeItem{
			ChangeID:             crontabChangeEvent.ChangeID,
			ChangeTimestampMicro: crontabChangeEvent.ChangeTimestampMicro,
			Title:                crontabChangeEvent.Title,
			Message:              crontabChangeEvent.Message,
		})
	}

	return changeEvents, nil
}

func (cc *CrontabChecker) getCrontabChange() (*Change[uint64, *Crontab], error) {
	// Get all crontab files to monitor
	crontabFiles, err := getCrontabFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get crontab files: %w", err)
	}

	change := NewChange[uint64, *Crontab]()

	for _, file := range crontabFiles {
		changeKey := GetHashCode([]byte(file))
		childChange := NewChange[uint64, *Crontab]()
		childChange.RawKey = file
		modTime, err := GetFileModTime(file)
		if err != nil {
			if !os.IsNotExist(err) {
				childChange.SetValid(false)
				l.Warnf("get file mod time failed: %s", err.Error())
				change.SetChild(changeKey, childChange)
			} else {
				l.Warnf("file %s not exist", file)
			}
			continue
		}

		// check if mod time changed
		if cc.crontabChange != nil {
			oldChange := cc.crontabChange.GetChild(changeKey)
			if oldChange != nil {
				if oldChange.LastUpdateTime.Equal(modTime) { // mod time not changed
					// mod time not changed
					change.SetChild(changeKey, oldChange)
					continue
				}
			}
		}

		content, contentHash, err := cc.getFileContentAndHash(file)
		if err != nil && !os.IsNotExist(err) {
			l.Warnf("get file content and hash failed: %s", err.Error())
			childChange.SetValid(false)
			change.SetChild(changeKey, childChange)
			continue
		}
		childChange.LastUpdateTime = modTime
		childChange.ContentHash = contentHash

		crontabs, err := ParseCrontab(file, content)
		if err != nil {
			l.Warnf("parse crontab failed: %s", err.Error())
			childChange.SetValid(false)
			change.SetChild(changeKey, childChange)
			continue
		}

		for _, crontab := range crontabs {
			key := getCrontabKey(crontab)
			childChange.Add(key, crontab)
		}

		change.SetChild(changeKey, childChange)
	}

	return change, nil
}

func (cc *CrontabChecker) getChangeEvent(file string, cronFileChanges *CronFileChanges) CrontabChangeItem {
	// Create merged change event
	changeItem := CrontabMergedChangeItem{
		ChangeID:             changes.ChangeID(ChangeIDCrontab),
		ChangeTimestampMicro: time.Now().UnixMicro(),
		ChangesByFile:        map[string]*CronFileChanges{file: cronFileChanges},
	}

	// Render title and message
	title, message, err := changes.RenderHostTemplate(defaultChangeLanguage, changeItem.ChangeID, changeItem)
	if err != nil {
		l.Warnf("RenderHostTemplate failed for %s: %v, using default message", string(changeItem.ChangeID), err.Error())
		// Fallback to default title and message
		title = "Crontab Job Changes"
		message = "Crontab files have been modified"
	}

	changeItem.Title = title
	changeItem.Message = message

	// Convert to CrontabChangeItem for compatibility with existing code
	convertedItem := CrontabChangeItem{
		ChangeID:             changeItem.ChangeID,
		ChangeTimestampMicro: changeItem.ChangeTimestampMicro,
		Title:                changeItem.Title,
		Message:              changeItem.Message,
	}

	return convertedItem
}

func (cc *CrontabChecker) getFileContentAndHash(file string) ([]byte, uint64, error) {
	content, err := ReadFile(file)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read %s: %w", file, err)
	}

	contentHash := GetHashCode(content)

	return content, contentHash, nil
}

// ParseCrontab parse crontab file
// This function can be tested with custom user crontab paths by modifying
// DirPathSpoolCrontabs variable before calling this function.
func ParseCrontab(crontabPath string, content []byte) ([]*Crontab, error) {
	lines := strings.Split(string(content), "\n")
	var crontabs []*Crontab

	// Determine if this is a user crontab file (from /var/spool/cron/crontabs/)
	isUserCrontab := strings.HasPrefix(crontabPath, DirPathSpoolCrontabs)

	// Extract username from user crontab file path
	username := ""
	if isUserCrontab {
		username = filepath.Base(crontabPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip environment variable lines (like PATH=/usr/bin:/bin)
		if !strings.Contains(line, " ") {
			continue
		}

		parts := strings.Fields(line)
		if isUserCrontab {
			// User crontab format: minute hour day month weekday command
			if len(parts) < 6 {
				// Skip invalid crontab entries
				continue
			}

			schedule := strings.Join(parts[0:5], " ")
			command := strings.Join(parts[5:], " ")

			crontab := &Crontab{
				User:     username,
				Schedule: schedule,
				Command:  command,
				FilePath: crontabPath,
			}

			crontabs = append(crontabs, crontab)
		} else {
			// System crontab format: minute hour day month weekday user command
			if len(parts) < 7 {
				// Skip invalid crontab entries
				continue
			}

			schedule := strings.Join(parts[0:5], " ")
			user := parts[5]
			command := strings.Join(parts[6:], " ")

			crontab := &Crontab{
				User:     user,
				Schedule: schedule,
				Command:  command,
				FilePath: crontabPath,
			}

			crontabs = append(crontabs, crontab)
		}
	}

	return crontabs, nil
}

// getCrontabKey generates a unique key for a crontab entry based on its schedule, user, command, and file path.
func getCrontabKey(cron *Crontab) uint64 {
	key := fmt.Sprintf("%s:%s:%s", cron.Schedule, cron.User, cron.Command)

	if cron.FilePath != "" {
		key = fmt.Sprintf("%s:%s:%s:%s", cron.Schedule, cron.User, cron.Command, cron.FilePath)
	}
	return GetHashCode([]byte(key))
}
