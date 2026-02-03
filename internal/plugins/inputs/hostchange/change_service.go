// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
)

const (
	// Service related change ID.
	ChangeIDCreateService changes.ChangeID = "host_change_04_01"
	ChangeIDDeleteService changes.ChangeID = "host_change_04_02"
	ChangeIDModifyService changes.ChangeID = "host_change_04_03"
	ChangeIDServiceStatus changes.ChangeID = "host_change_04_04"
)

// ServiceType represents the type of service management system.
type ServiceType string

func (st ServiceType) String() string {
	return string(st)
}

const (
	ServiceTypeSystemd  ServiceType = "systemd"
	ServiceTypeSysVinit ServiceType = "sysvinit"
	ServiceTypeUnknown  ServiceType = "unknown"
)

// ServiceStatus represents the status of a service.
type ServiceStatus string

const (
	ServiceStatusRunning      ServiceStatus = "running"
	ServiceStatusStopped      ServiceStatus = "stopped"
	ServiceStatusFailed       ServiceStatus = "failed"
	ServiceStatusActivating   ServiceStatus = "activating"
	ServiceStatusDeactivating ServiceStatus = "deactivating"
	ServiceStatusUnknown      ServiceStatus = "unknown"
)

// Service represents a system service.
type Service struct {
	Name        string        `json:"name"`         // Service name
	Type        ServiceType   `json:"type"`         // Service management type (systemd/sysvinit)
	Status      ServiceStatus `json:"status"`       // Current status
	Enabled     bool          `json:"enabled"`      // Whether the service is enabled on boot
	Path        string        `json:"path"`         // Path to service file (if applicable)
	Content     string        `json:"content"`      // Service file content (if applicable)
	ContentHash uint64        `json:"content_hash"` // Hash of service file content
	ModTime     int64         `json:"mod_time"`     // Modification time of service file
}

// ServiceChangeItem represents a service change event.
type ServiceChangeItem struct {
	ChangeID             changes.ChangeID `json:"change_id"`              // Unique change identifier
	ChangeTimestampMicro int64            `json:"change_timestamp_micro"` // Change timestamp in microseconds
	Title                string           `json:"title"`                  // Event title
	Message              string           `json:"message"`                // Detailed message
	ServiceName          string           `json:"service_name"`           // Service name
	ServiceType          ServiceType      `json:"service_type"`           // Service type
	OldStatus            ServiceStatus    `json:"old_status"`             // Old status
	NewStatus            ServiceStatus    `json:"new_status"`             // New status
	OldEnabled           bool             `json:"old_enabled"`            // Old enabled state
	NewEnabled           bool             `json:"new_enabled"`            // New enabled state
	ContentChanged       bool             `json:"content_changed"`        // Whether service file content changed
	DiffText             string           `json:"diff_text"`              // Difference text of service file content
}

type ServiceChangesByType map[string][]ServiceChangeItem

// ServiceChecker handles service change detection.
type ServiceChecker struct {
	Enabled         bool     `toml:"enabled"`          // Enable service change detection
	IgnoreServices  []string `toml:"ignore_services"`  // Services to ignore (supports regex)
	IncludeServices []string `toml:"include_services"` // Services to include (supports regex)
	ServiceTypes    []string `toml:"service_types"`    // Service types to monitor (systemd/sysvinit)

	input           *Input
	serviceChange   *Change[string, *Service]
	systemdEnabled  bool
	sysvinitEnabled bool
}

// Init initializes the ServiceChecker.
func (sc *ServiceChecker) Init(input *Input) error {
	sc.input = input

	// Detect available service management systems
	sc.detectServiceTypes()

	return nil
}

// Collect collects service changes and returns them as ChangeItem.
func (sc *ServiceChecker) Collect() ([]*ChangeItem, error) {
	l.Debugf("collecting service changes")

	changeEvents := []*ChangeItem{}
	changesByType := make(ServiceChangesByType)

	// Skip if no service management system is available
	if !sc.systemdEnabled && !sc.sysvinitEnabled {
		l.Warnf("no service management system available, skipping service change detection")
		return changeEvents, nil
	}

	// Get current services information
	currentServices, err := sc.getCurrentServices()
	if err != nil {
		return nil, fmt.Errorf("failed to get current services: %w", err)
	}

	newChange := NewChange[string, *Service]()

	for name, service := range currentServices {
		newChange.Add(name, service)
	}

	// Collect changes using GetChangeEvent
	if sc.serviceChange != nil {
		_, err := newChange.GetChangeEvent(sc.serviceChange, &ChangeItemConfig[string, *Service]{
			Add:    sc.add(changesByType),
			Delete: sc.delete(changesByType),
			Modify: sc.modify(changesByType),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get service change event: %w", err)
		}
	}

	// Convert collected changes to ChangeItem
	var allServiceChangeItems []ServiceChangeItem
	for _, changes := range changesByType {
		allServiceChangeItems = append(allServiceChangeItems, changes...)
	}

	// Render messages for all changes
	for _, value := range allServiceChangeItems {
		change := value
		changeItem, err := sc.renderServiceChangeMessage(&change)
		if err != nil {
			l.Warnf("render service %s change message failed: %s", change.ServiceName, err.Error())
			continue
		}
		changeEvents = append(changeEvents, changeItem)
	}

	// Update service change
	sc.serviceChange = newChange

	return changeEvents, nil
}

func (sc *ServiceChecker) add(changesByType ServiceChangesByType) GetChangeItemFunc[string, *Service] {
	return func(key string, newValues, oldValues []*Service, parentChanges ...*Change[string, *Service]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		service := newValues[0]
		currentTime := time.Now().UnixMicro()

		changeItem := ServiceChangeItem{
			ChangeID:             ChangeIDCreateService,
			ChangeTimestampMicro: currentTime,
			ServiceName:          service.Name,
			ServiceType:          service.Type,
			NewStatus:            service.Status,
			NewEnabled:           service.Enabled,
		}

		changeType := string(ChangeIDCreateService)
		changesByType[changeType] = append(changesByType[changeType], changeItem)

		return nil
	}
}

func (sc *ServiceChecker) delete(changesByType ServiceChangesByType) GetChangeItemFunc[string, *Service] {
	return func(key string, newValues, oldValues []*Service, parentChanges ...*Change[string, *Service]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		service := oldValues[0]
		currentTime := time.Now().UnixMicro()

		changeItem := ServiceChangeItem{
			ChangeID:             ChangeIDDeleteService,
			ChangeTimestampMicro: currentTime,
			ServiceName:          service.Name,
			ServiceType:          service.Type,
			OldStatus:            service.Status,
			OldEnabled:           service.Enabled,
		}

		changeType := string(ChangeIDDeleteService)
		changesByType[changeType] = append(changesByType[changeType], changeItem)

		return nil
	}
}

func (sc *ServiceChecker) modify(changesByType ServiceChangesByType) GetChangeItemFunc[string, *Service] {
	return func(key string, newValues, oldValues []*Service, parentChanges ...*Change[string, *Service]) *ChangeItem {
		if len(newValues) != 1 || len(oldValues) != 1 {
			return nil
		}

		newService := newValues[0]
		oldService := oldValues[0]
		currentTime := time.Now().UnixMicro()

		// Check for modifications
		var statusChanged bool
		var enabledChanged bool
		var contentChanged bool

		if oldService.Status != newService.Status {
			statusChanged = true
		}

		if oldService.Enabled != newService.Enabled {
			enabledChanged = true
		}

		if oldService.ContentHash != newService.ContentHash {
			contentChanged = true
		}

		if statusChanged {
			// Status change detected
			changeItem := ServiceChangeItem{
				ChangeID:             ChangeIDServiceStatus,
				ChangeTimestampMicro: currentTime,
				ServiceName:          newService.Name,
				ServiceType:          newService.Type,
				OldStatus:            oldService.Status,
				NewStatus:            newService.Status,
				OldEnabled:           oldService.Enabled,
				NewEnabled:           newService.Enabled,
			}

			changeType := string(ChangeIDServiceStatus)
			changesByType[changeType] = append(changesByType[changeType], changeItem)
		}

		if enabledChanged || contentChanged {
			// Calculate diff text if content changed
			diffText := ""
			if contentChanged {
				diffResult := diff.LineDiffWithContextLines(oldService.Content, newService.Content, 4)
				diffText = diffResult
			}

			// Service configuration changed
			changeItem := ServiceChangeItem{
				ChangeID:             ChangeIDModifyService,
				ChangeTimestampMicro: currentTime,
				ServiceName:          newService.Name,
				ServiceType:          newService.Type,
				OldStatus:            oldService.Status,
				NewStatus:            newService.Status,
				OldEnabled:           oldService.Enabled,
				NewEnabled:           newService.Enabled,
				ContentChanged:       contentChanged,
				DiffText:             diffText,
			}

			changeType := string(ChangeIDModifyService)
			changesByType[changeType] = append(changesByType[changeType], changeItem)
		}

		return nil
	}
}

// detectServiceTypes detects which service management systems are available and applies service_types configuration.
func (sc *ServiceChecker) detectServiceTypes() {
	sysvinitAvailable := false
	systemdAvailable := false

	// First detect what's available on the system
	_, systemdAvailableErr := exec.LookPath("systemctl")
	systemdAvailable = systemdAvailableErr == nil

	if _, err := os.Stat("/etc/init.d"); err == nil {
		sysvinitAvailable = true
	}

	// If service_types is not specified or empty, use all available service types
	if len(sc.ServiceTypes) == 0 {
		sc.systemdEnabled = systemdAvailable
		sc.sysvinitEnabled = sysvinitAvailable
	} else {
		// Otherwise, enable only the specified service types if they are available
		sc.systemdEnabled = false
		sc.sysvinitEnabled = false

		for _, serviceType := range sc.ServiceTypes {
			switch strings.ToLower(serviceType) {
			case ServiceTypeSystemd.String():
				sc.systemdEnabled = systemdAvailable
			case ServiceTypeSysVinit.String():
				sc.sysvinitEnabled = sysvinitAvailable
			default:
				l.Warnf("unknown service type: %s, ignoring", serviceType)
			}
		}
	}

	l.Infof("service types configured: systemd=%v, sysvinit=%v",
		sc.systemdEnabled, sc.sysvinitEnabled)
}

// getCurrentServices collects information about all current services.
func (sc *ServiceChecker) getCurrentServices() (map[string]*Service, error) {
	services := make(map[string]*Service)

	// Collect systemd services if available
	if sc.systemdEnabled {
		systemdServices, err := sc.getSystemdServices()
		if err != nil {
			l.Warnf("failed to get systemd services: %v", err)
		} else {
			for name, service := range systemdServices {
				services[name] = service
			}
		}
	}

	// Collect sysvinit services if available
	if sc.sysvinitEnabled {
		sysvinitServices, err := sc.getSysVinitServices()
		if err != nil {
			l.Warnf("failed to get sysvinit services: %v", err)
		} else {
			for name, service := range sysvinitServices {
				// Only add if not already added by systemd (which might manage sysvinit services too)
				if _, exists := services[name]; !exists {
					services[name] = service
				}
			}
		}
	}

	return services, nil
}

// getSystemdServices collects information about all systemd services.
func (sc *ServiceChecker) getSystemdServices() (map[string]*Service, error) {
	services := make(map[string]*Service)

	// Get list of all services
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-legend", "--output=json") //nolint:gosec
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list systemd services: %w", err)
	}

	// Parse the JSON output
	var units []map[string]interface{}
	if err := json.Unmarshal(output, &units); err != nil {
		return nil, fmt.Errorf("failed to parse systemd services JSON: %w", err)
	}

	for _, unit := range units {
		name := getString(unit, "unit", "Name", "Id")
		if !strings.HasSuffix(name, ".service") {
			continue
		}

		// Remove .service suffix
		serviceName := strings.TrimSuffix(name, ".service")

		// Check if service should be ignored
		if sc.shouldIgnoreService(serviceName) {
			continue
		}

		// Get service status with proper error handling
		status := ServiceStatusUnknown

		// Try both "activeState" and "active" fields for compatibility with different systemd versions
		stateStr := getString(unit, "active", "activeState")
		if stateStr != "" {
			status = sc.parseSystemdStatus(stateStr)
		}

		// Check if service is enabled
		enabled, err := sc.isSystemdServiceEnabled(serviceName)
		if err != nil {
			l.Debugf("failed to check if service %s is enabled: %v", serviceName, err)
			enabled = false
		}

		// Get service file path and content
		filePath, content, modTime, contentHash, err := sc.getSystemdServiceFileInfo(serviceName)
		if err != nil {
			l.Warnf("failed to get service file info for %s: %v", serviceName, err)
			continue
		}

		services[serviceName] = &Service{
			Name:        serviceName,
			Type:        ServiceTypeSystemd,
			Status:      status,
			Enabled:     enabled,
			Path:        filePath,
			Content:     content,
			ContentHash: contentHash,
			ModTime:     modTime,
		}
	}

	return services, nil
}

// isSystemdServiceEnabled checks if a systemd service is enabled.
func (sc *ServiceChecker) isSystemdServiceEnabled(name string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", name+".service") //nolint:gosec
	output, err := cmd.Output()
	if err != nil {
		// Command returns non-zero exit code if not enabled
		return false, nil
	}

	return strings.TrimSpace(string(output)) == "enabled", nil
}

// getSystemdServiceFileInfo gets information about a systemd service file.
func (sc *ServiceChecker) getSystemdServiceFileInfo(name string) (string, string, int64, uint64, error) {
	// Get service file path
	cmd := exec.Command("systemctl", "show", name+".service", "--property=FragmentPath") //nolint:gosec
	output, err := cmd.Output()
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to get service file path: %w", err)
	}

	path := strings.TrimPrefix(strings.TrimSpace(string(output)), "FragmentPath=")
	if path == "" {
		l.Debugf("service [%s] file path is empty, ignore reading content", name)
		return "", "", 0, 0, nil
	}

	// Read service file content
	content, err := ReadFile(path)
	if err != nil {
		return path, "", 0, 0, fmt.Errorf("failed to read service file: %w", err)
	}

	// Get file modification time
	info, err := os.Stat(path)
	if err != nil {
		return path, string(content), 0, GetHashCode(content), fmt.Errorf("failed to get file info: %w", err)
	}

	return path, string(content), info.ModTime().Unix(), GetHashCode(content), nil
}

// parseSystemdStatus converts systemd status string to ServiceStatus.
func (sc *ServiceChecker) parseSystemdStatus(status string) ServiceStatus {
	switch strings.ToLower(status) {
	case "active":
		return ServiceStatusRunning
	case "inactive":
		return ServiceStatusStopped
	case "failed":
		return ServiceStatusFailed
	case "activating":
		return ServiceStatusActivating
	case "deactivating":
		return ServiceStatusDeactivating
	default:
		return ServiceStatusUnknown
	}
}

// sysvinitServiceRegex is a precompiled regex to extract service names from rc.d files.
var sysvinitServiceRegex = regexp.MustCompile(`^S\d+([a-zA-Z0-9_-]+)$`)

// getEnabledSysVinitServices scans runlevel directories to find enabled services.
func (sc *ServiceChecker) getEnabledSysVinitServices() map[string]bool {
	enabledServices := make(map[string]bool)
	// Scan all runlevel directories from rc2.d to rc5.d
	for runlevel := 2; runlevel <= 5; runlevel++ {
		runlevelDir := fmt.Sprintf("/etc/rc%d.d", runlevel)
		files, err := os.ReadDir(runlevelDir)
		if err != nil {
			continue // Skip if runlevel directory doesn't exist
		}

		for _, file := range files {
			name := file.Name()
			// Only process files starting with 'S' (start)
			if strings.HasPrefix(name, "S") {
				// Extract service name using precompiled regex
				match := sysvinitServiceRegex.FindStringSubmatch(name)
				if len(match) > 1 {
					serviceName := match[1]
					enabledServices[serviceName] = true
				}
			}
		}
	}
	return enabledServices
}

// getSysVinitServices collects information about all sysvinit services.
func (sc *ServiceChecker) getSysVinitServices() (map[string]*Service, error) {
	services := make(map[string]*Service)

	// Precompute enabled services map to avoid repeated directory reads
	enabledServices := sc.getEnabledSysVinitServices()

	// Read all files in /etc/init.d/
	files, err := os.ReadDir("/etc/init.d")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/init.d directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		serviceName := file.Name()

		// Check if service should be ignored
		if sc.shouldIgnoreService(serviceName) {
			continue
		}

		// Check if file is executable
		path := filepath.Join("/etc/init.d", serviceName)
		info, err := file.Info()
		if err != nil {
			l.Debugf("failed to get info for %s: %v", path, err)
			continue
		}

		if info.Mode()&0o111 == 0 {
			// Not executable, probably not a service
			continue
		}

		// Get service status
		status, err := sc.getSysVinitServiceStatus(serviceName)
		if err != nil {
			l.Debugf("failed to get status for service %s: %v", serviceName, err)
			status = ServiceStatusUnknown
		}

		// Check if service is enabled from precomputed map
		enabled := enabledServices[serviceName]

		// Read service file content
		content, err := ReadFile(path)
		if err != nil {
			l.Debugf("failed to read service file %s: %v", path, err)
		}

		services[serviceName] = &Service{
			Name:        serviceName,
			Type:        ServiceTypeSysVinit,
			Status:      status,
			Enabled:     enabled,
			Path:        path,
			Content:     string(content),
			ContentHash: GetHashCode(content),
			ModTime:     info.ModTime().Unix(),
		}
	}

	return services, nil
}

// getSysVinitServiceStatus gets the status of a sysvinit service.
func (sc *ServiceChecker) getSysVinitServiceStatus(name string) (ServiceStatus, error) {
	cmd := exec.Command(filepath.Join("/etc/init.d", name), "status") //nolint:gosec
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Check for common status indicators
	switch {
	case isContainString(outputStr, "is running", "running"):
		return ServiceStatusRunning, nil
	case isContainString(outputStr, "is stopped", "stopped", "dead"):
		return ServiceStatusStopped, nil
	case isContainString(outputStr, "failed", "FAILED"):
		return ServiceStatusFailed, nil
	}

	// If command failed but we can't determine status
	if err != nil {
		return ServiceStatusUnknown, nil
	}

	return ServiceStatusUnknown, nil
}

// shouldIgnoreService checks if a service should be ignored.
func (sc *ServiceChecker) shouldIgnoreService(name string) bool {
	// If include_services is not empty, service must match at least one pattern
	if len(sc.IncludeServices) > 0 {
		includeMatched := false
		for _, include := range sc.IncludeServices {
			if include == name {
				includeMatched = true
				break
			}
			// Try regex match
			if matched, err := regexp.MatchString(include, name); err == nil && matched {
				includeMatched = true
				break
			}
		}
		if !includeMatched {
			return true
		}
	}

	// Check if service should be ignored
	for _, ignore := range sc.IgnoreServices {
		if ignore == name {
			return true
		}
		// Try regex match
		if matched, err := regexp.MatchString(ignore, name); err == nil && matched {
			return true
		}
		// Fallback to wildcard match for backward compatibility
		if matched, _ := filepath.Match(ignore, name); matched {
			return true
		}
	}

	return false
}

// renderServiceChangeMessage renders a message for a service change using templates.
func (sc *ServiceChecker) renderServiceChangeMessage(change *ServiceChangeItem) (*ChangeItem, error) {
	// Prepare data for template rendering
	data := struct {
		ServiceName    string        `json:"service_name"`
		ServiceType    ServiceType   `json:"service_type"`
		OldStatus      ServiceStatus `json:"old_status"`
		NewStatus      ServiceStatus `json:"new_status"`
		OldEnabled     bool          `json:"old_enabled"`
		NewEnabled     bool          `json:"new_enabled"`
		EnabledChanged bool          `json:"enabled_changed"`
		ContentChanged bool          `json:"content_changed"`
		DiffText       string        `json:"diff_text"`
	}{
		ServiceName:    change.ServiceName,
		ServiceType:    change.ServiceType,
		OldStatus:      change.OldStatus,
		NewStatus:      change.NewStatus,
		OldEnabled:     change.OldEnabled,
		NewEnabled:     change.NewEnabled,
		EnabledChanged: change.OldEnabled != change.NewEnabled,
		ContentChanged: change.ContentChanged,
		DiffText:       change.DiffText,
	}

	// Render template using host.toml
	title, message, err := changes.RenderHostTemplate(defaultChangeLanguage, change.ChangeID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render host template: %w", err)
	}

	return &ChangeItem{
		ChangeID:             change.ChangeID,
		ChangeTimestampMicro: change.ChangeTimestampMicro,
		Title:                title,
		Message:              message,
	}, nil
}
