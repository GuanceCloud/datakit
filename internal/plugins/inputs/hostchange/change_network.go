// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
)

// Network configuration related change ID.
const (
	ChangeIDNetworkInterface changes.ChangeID = "host_change_05_01" // Network interface change (includes interface status and IP address changes)
	ChangeIDNetworkDNS       changes.ChangeID = "host_change_05_02" // DNS configuration change
	ChangeIDNetworkRoute     changes.ChangeID = "host_change_05_03" // Route configuration change
	ChangeIDNetworkFirewall  changes.ChangeID = "host_change_05_04" // Firewall rule change
	ChangeIDNetworkHosts     changes.ChangeID = "host_change_05_05" // Hosts file change
)

var iptablesCounterRegex = regexp.MustCompile(`\[\d+:\d+\]`) // Regex to match iptables packet/byte counters [packet_count:byte_count]

// NetworkInterface represents a network interface configuration.
type NetworkInterface struct {
	Name      string   `json:"name"`      // Interface name
	Type      string   `json:"type"`      // Interface type (ethernet, wifi, bridge, etc.)
	MAC       string   `json:"mac"`       // MAC address
	MTU       int      `json:"mtu"`       // MTU value
	IPs       []string `json:"ips"`       // IP address list
	Gateway   string   `json:"gateway"`   // Gateway address
	Netmask   string   `json:"netmask"`   // Netmask
	Broadcast string   `json:"broadcast"` // Broadcast address
	Up        bool     `json:"up"`        // Interface status
	FilePath  string   `json:"file_path"` // Configuration file path
}

// DNSConfig represents DNS configuration.
type DNSConfig struct {
	Nameservers []string `json:"nameservers"` // DNS server list
	Search      []string `json:"search"`      // Search domain list
	FilePath    string   `json:"file_path"`   // Configuration file path
}

// RouteConfig represents route configuration.
type RouteConfig struct {
	Destination string `json:"destination"` // Destination network
	Gateway     string `json:"gateway"`     // Gateway address
	Genmask     string `json:"genmask"`     // Netmask
	Flags       string `json:"flags"`       // Route flags
	Metric      int    `json:"metric"`      // Route metric
	Ref         int    `json:"ref"`         // Reference count
	Use         int    `json:"use"`         // Use count
	Iface       string `json:"iface"`       // Output interface
}

// FirewallRule represents a firewall rule.
type FirewallRule struct {
	Chain       string `json:"chain"`       // Chain name
	Protocol    string `json:"protocol"`    // Protocol
	Source      string `json:"source"`      // Source address
	Destination string `json:"destination"` // Destination address
	Sport       string `json:"sport"`       // Source port
	Dport       string `json:"dport"`       // Destination port
	Action      string `json:"action"`      // Action
}

// HostsEntry represents a hosts file entry.
type HostsEntry struct {
	Hostname string   `json:"hostname"` // Hostname (used as key)
	IPs      []string `json:"ips"`      // IP addresses
}

// NetworkConfigChangeItem represents a network configuration change event.
type NetworkConfigChangeItem struct {
	ChangeID             changes.ChangeID `json:"change_id"`              // Unique change identifier
	ChangeTimestampMicro int64            `json:"change_timestamp_micro"` // Change timestamp in microseconds
	Title                string           `json:"title"`                  // Event title
	Message              string           `json:"message"`                // Detailed message
	InterfaceName        string           `json:"interface_name"`         // Network interface name
	OldIP                string           `json:"old_ip"`                 // Old IP address
	NewIP                string           `json:"new_ip"`                 // New IP address
	OldStatus            bool             `json:"old_status"`             // Old interface status
	NewStatus            bool             `json:"new_status"`             // New interface status
	ConfigType           string           `json:"config_type"`            // Configuration type (interface, dns, route, firewall)
}

// NetworkConfigChecker handles network configuration change detection.
type NetworkConfigChecker struct {
	Enabled          bool     `toml:"enabled"`           // Enable network configuration change detection
	IgnoreInterfaces []string `toml:"ignore_interfaces"` // Interfaces to ignore

	input          *Input
	networkChange  *Change[string, *NetworkInterface]
	dnsChange      *Change[string, string]
	routeChange    *Change[string, string]
	firewallChange *Change[string, string]
	hostsChange    *Change[string, string]
}

// Init initializes the NetworkConfigChecker.
func (nc *NetworkConfigChecker) Init(input *Input) error {
	nc.input = input
	return nil
}

// Collect collects network configuration changes and returns them as ChangeItem.
func (nc *NetworkConfigChecker) Collect() ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	// 1. Get current network interfaces
	currentInterfaces, err := nc.getCurrentNetworkInterfaces()
	if err != nil {
		l.Warnf("failed to get current network interfaces: %s", err.Error())
	} else {
		// Create new changes
		newInterfaceChange := NewChange[string, *NetworkInterface]()
		for name, iface := range currentInterfaces {
			newInterfaceChange.Add(name, iface)
		}

		// Get change events
		if interfaceEvents, changeErr := nc.getInterfaceChangeEvents(newInterfaceChange); changeErr != nil {
			l.Warnf("failed to get interface change events: %s", changeErr.Error())
		} else {
			changeEvents = append(changeEvents, interfaceEvents...)
		}

		nc.networkChange = newInterfaceChange
	}

	// 2. Get current DNS configurations
	newDNSChange, err := nc.getCurrentDNSChange()
	if err != nil {
		l.Warnf("failed to get current DNS change: %s", err.Error())
	} else if dnsEvents, changeErr := nc.getDNSChangeEvents(newDNSChange); changeErr != nil {
		l.Warnf("failed to get DNS change events: %s", changeErr.Error())
	} else {
		changeEvents = append(changeEvents, dnsEvents...)
		nc.dnsChange = newDNSChange
	}

	// 3. Get current routes
	newRouteChange, err := nc.getCurrentRoutes()
	if err != nil {
		l.Warnf("failed to get current routes: %s", err.Error())
	} else if routeEvents, changeErr := nc.getRouteChangeEvents(newRouteChange); changeErr != nil {
		l.Warnf("failed to get route change events: %s", changeErr.Error())
	} else {
		changeEvents = append(changeEvents, routeEvents...)
		nc.routeChange = newRouteChange
	}

	// 4. Get current firewall rules
	newFirewallChange, err := nc.getCurrentFirewallRules()
	if err != nil {
		l.Warnf("failed to get current firewall rules: %s", err.Error())
	} else if firewallEvents, changeErr := nc.getFirewallChangeEvents(newFirewallChange); changeErr != nil {
		l.Warnf("failed to get firewall change events: %s", changeErr.Error())
	} else {
		changeEvents = append(changeEvents, firewallEvents...)
		nc.firewallChange = newFirewallChange
	}

	// 5. Get current hosts file entries
	newHostsChange, err := nc.getCurrentHostsChange()
	if err != nil {
		l.Warnf("failed to get current hosts file entries: %s", err.Error())
	} else if hostsEvents, changeErr := nc.getHostsChangeEvents(newHostsChange); changeErr != nil {
		l.Warnf("failed to get hosts change events: %s", changeErr.Error())
	} else {
		changeEvents = append(changeEvents, hostsEvents...)
		nc.hostsChange = newHostsChange
	}

	return changeEvents, nil
}

// getCurrentNetworkInterfaces collects information about all current network interfaces.
func (nc *NetworkConfigChecker) getCurrentNetworkInterfaces() (map[string]*NetworkInterface, error) {
	interfaces := make(map[string]*NetworkInterface)

	// Get network interfaces from ip command
	cmd := exec.Command("ip", "addr")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w, output: %s", err, string(output))
	}

	// Parse ip addr output
	lines := strings.Split(string(output), "\n")
	var currentIface *NetworkInterface

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for interface line (e.g., "1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000")
		if strings.Contains(line, ":") && !strings.Contains(line, ":inet") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}

			name := strings.TrimSpace(parts[1])
			// Skip ignored interfaces
			if nc.shouldIgnoreInterface(name) {
				currentIface = nil
				continue
			}

			// Create new interface
			currentIface = &NetworkInterface{
				Name: name,
				IPs:  []string{},
				Up:   strings.Contains(line, "state UP"),
			}
			interfaces[name] = currentIface

			// Extract MAC address if present
			macIndex := strings.Index(line, "link/ether")
			if macIndex != -1 {
				macParts := strings.SplitN(line[macIndex:], " ", 3)
				if len(macParts) > 1 {
					currentIface.MAC = strings.TrimSpace(macParts[1])
				}
			}

			// Extract MTU if present
			mtuIndex := strings.Index(line, "mtu")
			if mtuIndex != -1 {
				mtuParts := strings.SplitN(line[mtuIndex:], " ", 3)
				if len(mtuParts) > 1 {
					if _, err := fmt.Sscanf(mtuParts[1], "%d", &currentIface.MTU); err != nil {
						l.Warnf("failed to parse MTU value: %s", err.Error())
					}
				}
			}
		} else if currentIface != nil {
			// Check for IP address line (e.g., "    inet 127.0.0.1/8 scope host lo")
			if strings.HasPrefix(line, "inet ") {
				ipParts := strings.SplitN(line, " ", 3)
				if len(ipParts) > 1 {
					ipAddr := strings.SplitN(ipParts[1], "/", 2)[0]
					currentIface.IPs = append(currentIface.IPs, ipAddr)
				}
			}
		}
	}

	return interfaces, nil
}

// getCurrentDNSChange collects current DNS configuration.
func (nc *NetworkConfigChecker) getCurrentDNSChange() (*Change[string, string], error) {
	newDNSChange := NewChange[string, string]()
	// Read resolv.conf file
	content, err := ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read resolv.conf: %w", err)
	}

	// Parse resolv.conf
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "nameserver":
			newDNSChange.Add("nameserver", parts[1])
		case "search":
			newDNSChange.Add("search", strings.Join(parts[1:], " "))
		case "options":
			newDNSChange.Add("options", strings.Join(parts[1:], " "))
		}
	}

	return newDNSChange, nil
}

// getCurrentRoutes collects current route configuration.
func (nc *NetworkConfigChecker) getCurrentRoutes() (*Change[string, string], error) {
	newRouteChange := NewChange[string, string]()

	// Get routes from ip route command
	cmd := exec.Command("ip", "route")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w, output: %s", err, string(output))
	}

	// Parse ip route output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// Parse route line (e.g., "default via 192.168.1.1 dev eth0 proto dhcp metric 100")
		destination := parts[0]

		// Extract route details excluding destination
		var routeDetails []string
		for i := 1; i < len(parts); i++ {
			routeDetails = append(routeDetails, parts[i])
		}

		// Add route with destination as key and details as value
		newRouteChange.Add(destination, strings.Join(routeDetails, " "))
	}

	return newRouteChange, nil
}

// getCurrentFirewallRules collects current firewall rules.
func (nc *NetworkConfigChecker) getCurrentFirewallRules() (*Change[string, string], error) {
	newFirewallChange := NewChange[string, string]()

	// Check if iptables is available
	_, err := exec.LookPath("iptables")
	if err != nil {
		// iptables not available, return empty rules
		return newFirewallChange, nil
	}

	// Get iptables rules using iptables-save
	cmd := exec.Command("iptables-save")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get iptables rules: %w, output: %s", err, string(output))
	}

	// Parse iptables-save output
	lines := strings.Split(string(output), "\n")
	var currentTable string
	var tableRules []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for table declaration (e.g., "*filter", "*nat", "*mangle", "*raw")
		if strings.HasPrefix(line, "*") {
			// Save previous table rules if any
			if currentTable != "" && len(tableRules) > 0 {
				rulesText := strings.Join(tableRules, "\n")
				newFirewallChange.Add(currentTable, rulesText)
			}
			// Start new table
			currentTable = strings.TrimPrefix(line, "*")
			tableRules = []string{}
			continue
		}

		// Check for COMMIT line (end of table)
		if line == "COMMIT" {
			// Save table rules
			if currentTable != "" && len(tableRules) > 0 {
				rulesText := strings.Join(tableRules, "\n")
				newFirewallChange.Add(currentTable, rulesText)
			}
			currentTable = ""
			tableRules = []string{}
			continue
		}

		// Collect rules for current table
		if currentTable != "" {
			// Remove packet/byte counters from rules to avoid false positives
			// Counters format: [packet_count:byte_count] e.g., [5866097:6957621656]
			cleanedLine := iptablesCounterRegex.ReplaceAllString(line, "")
			tableRules = append(tableRules, cleanedLine)
		}
	}

	// Save last table if not committed
	if currentTable != "" && len(tableRules) > 0 {
		rulesText := strings.Join(tableRules, "\n")
		newFirewallChange.Add(currentTable, rulesText)
	}

	return newFirewallChange, nil
}

// getCurrentHostsChange collects current hosts file entries.
func (nc *NetworkConfigChecker) getCurrentHostsChange() (*Change[string, string], error) {
	newHostsChange := NewChange[string, string]()

	content, err := ReadFile("/etc/hosts")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/hosts: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		ip := parts[0]
		hostnames := parts[1:]

		for _, hostname := range hostnames {
			if strings.HasPrefix(hostname, "#") {
				break
			}

			newHostsChange.Add(hostname, ip)
		}
	}

	return newHostsChange, nil
}

// shouldIgnoreInterface checks if a network interface should be ignored.
func (nc *NetworkConfigChecker) shouldIgnoreInterface(name string) bool {
	for _, ignore := range nc.IgnoreInterfaces {
		if ignore == name {
			return true
		}
		// Try wildcard match
		if matched, _ := filepath.Match(ignore, name); matched {
			return true
		}
	}
	return false
}

// addInterfaceChange handles interface addition changes.
func (nc *NetworkConfigChecker) addInterfaceChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, *NetworkInterface] {
	return func(key string, newValues, oldValues []*NetworkInterface, parentChanges ...*Change[string, *NetworkInterface]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		newIface := newValues[0]
		newIP := "-"
		if len(newIface.IPs) > 0 {
			newIP = newIface.IPs[0]
		}
		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkInterface,
			IfaceName:  key,
			OldIP:      "-",
			NewIP:      newIP,
			OldStatus:  false,
			NewStatus:  newIface.Up,
			ConfigType: "interface",
			ChangeType: "add",
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// deleteInterfaceChange handles interface deletion changes.
func (nc *NetworkConfigChecker) deleteInterfaceChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, *NetworkInterface] {
	return func(key string, newValues, oldValues []*NetworkInterface, parentChanges ...*Change[string, *NetworkInterface]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		oldIface := oldValues[0]
		oldIP := "-"
		if len(oldIface.IPs) > 0 {
			oldIP = oldIface.IPs[0]
		}
		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkInterface,
			IfaceName:  key,
			OldIP:      oldIP,
			NewIP:      "-",
			OldStatus:  oldIface.Up,
			NewStatus:  false,
			ConfigType: "interface",
			ChangeType: "delete",
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// modifyInterfaceChange handles interface modification changes.
func (nc *NetworkConfigChecker) modifyInterfaceChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, *NetworkInterface] {
	return func(key string, newValues, oldValues []*NetworkInterface, parentChanges ...*Change[string, *NetworkInterface]) *ChangeItem {
		if len(newValues) != 1 || len(oldValues) != 1 {
			return nil
		}

		newIface := newValues[0]
		oldIface := oldValues[0]

		// Status change
		if oldIface.Up != newIface.Up {
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkInterface,
				IfaceName:  key,
				OldIP:      "-",
				NewIP:      "-",
				OldStatus:  oldIface.Up,
				NewStatus:  newIface.Up,
				ConfigType: "interface",
				ChangeType: "modify",
			})
			*changeItems = append(*changeItems, changeItem)
		}

		// IP change
		if !compareStringSlices(oldIface.IPs, newIface.IPs) {
			oldIP := ""
			newIP := ""
			if len(oldIface.IPs) > 0 {
				oldIP = strings.Join(oldIface.IPs, ",")
			}
			if len(newIface.IPs) > 0 {
				newIP = strings.Join(newIface.IPs, ",")
			}
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkInterface,
				IfaceName:  key,
				OldIP:      oldIP,
				NewIP:      newIP,
				OldStatus:  oldIface.Up,
				NewStatus:  newIface.Up,
				ConfigType: "ip",
				ChangeType: "modify",
			})
			*changeItems = append(*changeItems, changeItem)
		}

		return nil
	}
}

// getInterfaceChangeEvents generates change events for network interfaces.
func (nc *NetworkConfigChecker) getInterfaceChangeEvents(newChange *Change[string, *NetworkInterface]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	if nc.networkChange == nil {
		// First run, no previous data
		return changeEvents, nil
	}

	// Collect changes using GetChangeEvent
	_, err := newChange.GetChangeEvent(nc.networkChange, &ChangeItemConfig[string, *NetworkInterface]{
		Add:    nc.addInterfaceChange(&changeEvents),
		Delete: nc.deleteInterfaceChange(&changeEvents),
		Modify: nc.modifyInterfaceChange(&changeEvents),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get interface change event: %w", err)
	}

	return changeEvents, nil
}

// addDNSChange handles DNS addition changes.
func (nc *NetworkConfigChecker) addDNSChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		diffText := strings.Join(newValues, "\n")
		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkDNS,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "dns",
			ChangeType: "add",
			DiffText:   diffText,
			DiffKey:    key,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// deleteDNSChange handles DNS deletion changes.
func (nc *NetworkConfigChecker) deleteDNSChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		diffText := strings.Join(oldValues, "\n")

		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkDNS,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "dns",
			ChangeType: "delete",
			DiffKey:    key,
			DiffText:   diffText,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// modifyDNSChange handles DNS modification changes.
func (nc *NetworkConfigChecker) modifyDNSChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 || len(oldValues) == 0 {
			return nil
		}

		newDNS := strings.Join(newValues, "\n")
		oldDNS := strings.Join(oldValues, "\n")

		// Check for DNS changes
		if newDNS != oldDNS {
			diffText := diff.LineDiffWithContextLines(oldDNS, newDNS, 3)
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkDNS,
				IfaceName:  "",
				OldIP:      "",
				NewIP:      "",
				OldStatus:  false,
				NewStatus:  false,
				ConfigType: "dns",
				ChangeType: "modify",
				DiffText:   diffText,
				DiffKey:    key,
			})
			*changeItems = append(*changeItems, changeItem)
		}

		return nil
	}
}

// getDNSChangeEvents generates change events for DNS configuration.
func (nc *NetworkConfigChecker) getDNSChangeEvents(newChange *Change[string, string]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	if nc.dnsChange == nil {
		// First run, no previous data
		return changeEvents, nil
	}

	// Collect changes using GetChangeEvent
	_, err := newChange.GetChangeEvent(nc.dnsChange, &ChangeItemConfig[string, string]{
		Add:    nc.addDNSChange(&changeEvents),
		Delete: nc.deleteDNSChange(&changeEvents),
		Modify: nc.modifyDNSChange(&changeEvents),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS change event: %w", err)
	}

	return changeEvents, nil
}

// addRouteChange handles route addition changes.
func (nc *NetworkConfigChecker) addRouteChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		diffText := strings.Join(newValues, "\n")
		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkRoute,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "route",
			ChangeType: "add",
			DiffText:   diffText,
			DiffKey:    key,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// deleteRouteChange handles route deletion changes.
func (nc *NetworkConfigChecker) deleteRouteChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		diffText := strings.Join(oldValues, "\n")

		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkRoute,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "route",
			ChangeType: "delete",
			DiffKey:    key,
			DiffText:   diffText,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// modifyRouteChange handles route modification changes.
func (nc *NetworkConfigChecker) modifyRouteChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 || len(oldValues) == 0 {
			return nil
		}

		newRoute := strings.Join(newValues, "\n")
		oldRoute := strings.Join(oldValues, "\n")

		// Check for route changes
		if newRoute != oldRoute {
			diffText := diff.LineDiffWithContextLines(oldRoute, newRoute, 3)
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkRoute,
				IfaceName:  "",
				OldIP:      "",
				NewIP:      "",
				OldStatus:  false,
				NewStatus:  false,
				ConfigType: "route",
				ChangeType: "modify",
				DiffText:   diffText,
				DiffKey:    key,
			})
			*changeItems = append(*changeItems, changeItem)
		}

		return nil
	}
}

// getRouteChangeEvents generates change events for route configuration.
func (nc *NetworkConfigChecker) getRouteChangeEvents(newChange *Change[string, string]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	if nc.routeChange == nil {
		// First run, no previous data
		return changeEvents, nil
	}

	// Collect changes using GetChangeEvent
	_, err := newChange.GetChangeEvent(nc.routeChange, &ChangeItemConfig[string, string]{
		Add:    nc.addRouteChange(&changeEvents),
		Delete: nc.deleteRouteChange(&changeEvents),
		Modify: nc.modifyRouteChange(&changeEvents),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get route change event: %w", err)
	}

	return changeEvents, nil
}

// addFirewallChange handles firewall rule addition changes.
func (nc *NetworkConfigChecker) addFirewallChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		diffText := strings.Join(newValues, "\n")
		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkFirewall,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "firewall",
			ChangeType: "add",
			DiffText:   diffText,
			DiffKey:    key,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// deleteFirewallChange handles firewall rule deletion changes.
func (nc *NetworkConfigChecker) deleteFirewallChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		diffText := strings.Join(oldValues, "\n")

		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkFirewall,
			IfaceName:  "",
			OldIP:      "",
			NewIP:      "",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "firewall",
			ChangeType: "delete",
			DiffKey:    key,
			DiffText:   diffText,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// modifyFirewallChange handles firewall rule modification changes.
func (nc *NetworkConfigChecker) modifyFirewallChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 || len(oldValues) == 0 {
			return nil
		}

		newRules := strings.Join(newValues, "\n")
		oldRules := strings.Join(oldValues, "\n")

		// Check for firewall rule changes
		if newRules != oldRules {
			diffText := diff.LineDiffWithContextLines(oldRules, newRules, 3)
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkFirewall,
				IfaceName:  "",
				OldIP:      "",
				NewIP:      "",
				OldStatus:  false,
				NewStatus:  false,
				ConfigType: "firewall",
				ChangeType: "modify",
				DiffText:   diffText,
				DiffKey:    key,
			})
			*changeItems = append(*changeItems, changeItem)
		}

		return nil
	}
}

// getFirewallChangeEvents generates change events for firewall rules.
func (nc *NetworkConfigChecker) getFirewallChangeEvents(newChange *Change[string, string]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	if nc.firewallChange == nil {
		// First run, no previous data
		return changeEvents, nil
	}

	// Collect changes using GetChangeEvent
	_, err := newChange.GetChangeEvent(nc.firewallChange, &ChangeItemConfig[string, string]{
		Add:    nc.addFirewallChange(&changeEvents),
		Delete: nc.deleteFirewallChange(&changeEvents),
		Modify: nc.modifyFirewallChange(&changeEvents),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall change event: %w", err)
	}

	return changeEvents, nil
}

// addHostsChange handles hosts entry addition changes.
func (nc *NetworkConfigChecker) addHostsChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		newIP := newValues[0]

		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkHosts,
			IfaceName:  key,
			OldIP:      "-",
			NewIP:      newIP,
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "hosts",
			ChangeType: "add",
			DiffKey:    key,
			DiffText:   newIP,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// deleteHostsChange handles hosts entry deletion changes.
func (nc *NetworkConfigChecker) deleteHostsChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		oldIP := oldValues[0]

		changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
			ChangeID:   ChangeIDNetworkHosts,
			IfaceName:  key,
			OldIP:      oldIP,
			NewIP:      "-",
			OldStatus:  false,
			NewStatus:  false,
			ConfigType: "hosts",
			ChangeType: "delete",
			DiffKey:    key,
			DiffText:   oldIP,
		})
		*changeItems = append(*changeItems, changeItem)
		return nil
	}
}

// modifyHostsChange handles hosts entry modification changes.
func (nc *NetworkConfigChecker) modifyHostsChange(changeItems *[]*ChangeItem) GetChangeItemFunc[string, string] {
	return func(key string, newValues, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
		if len(newValues) == 0 || len(oldValues) == 0 {
			return nil
		}

		newIP := strings.Join(newValues, "\n")
		oldIP := strings.Join(oldValues, "\n")

		if newIP != oldIP {
			diffText := diff.LineDiffWithContextLines(oldIP, newIP, 3)
			changeItem := nc.createNetworkChangeItem(NetworkChangeParams{
				ChangeID:   ChangeIDNetworkHosts,
				IfaceName:  key,
				OldIP:      oldIP,
				NewIP:      newIP,
				OldStatus:  false,
				NewStatus:  false,
				ConfigType: "hosts",
				ChangeType: "modify",
				DiffKey:    key,
				DiffText:   diffText,
			})
			*changeItems = append(*changeItems, changeItem)
		}

		return nil
	}
}

// getHostsChangeEvents generates change events for hosts file.
func (nc *NetworkConfigChecker) getHostsChangeEvents(newChange *Change[string, string]) ([]*ChangeItem, error) {
	changeEvents := []*ChangeItem{}

	if nc.hostsChange == nil {
		return changeEvents, nil
	}

	_, err := newChange.GetChangeEvent(nc.hostsChange, &ChangeItemConfig[string, string]{
		Add:    nc.addHostsChange(&changeEvents),
		Delete: nc.deleteHostsChange(&changeEvents),
		Modify: nc.modifyHostsChange(&changeEvents),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts change event: %w", err)
	}

	return changeEvents, nil
}

// NetworkChangeParams holds the parameters for creating a network change item.
type NetworkChangeParams struct {
	ChangeID   changes.ChangeID
	IfaceName  string
	OldIP      string
	NewIP      string
	OldStatus  bool
	NewStatus  bool
	ConfigType string
	ChangeType string
	DiffText   string
	DiffKey    string
}

// createNetworkChangeItem creates a network configuration change item.
func (nc *NetworkConfigChecker) createNetworkChangeItem(params NetworkChangeParams) *ChangeItem {
	currentTime := time.Now().UnixMicro()

	// Prepare data for template rendering
	data := struct {
		InterfaceName string `json:"interface_name"`
		OldIP         string `json:"old_ip"`
		NewIP         string `json:"new_ip"`
		OldStatus     bool   `json:"old_status"`
		NewStatus     bool   `json:"new_status"`
		ConfigType    string `json:"config_type"`
		ChangeType    string `json:"change_type"`
		DiffText      string `json:"diff_text"`
		DiffKey       string `json:"diff_key"`
	}{
		InterfaceName: params.IfaceName,
		OldIP:         params.OldIP,
		NewIP:         params.NewIP,
		OldStatus:     params.OldStatus,
		NewStatus:     params.NewStatus,
		ConfigType:    params.ConfigType,
		ChangeType:    params.ChangeType,
		DiffText:      params.DiffText,
		DiffKey:       params.DiffKey,
	}

	// Render template
	title, message, err := changes.RenderHostTemplate(defaultChangeLanguage, params.ChangeID, data)
	if err != nil {
		l.Warnf("failed to render network change template: %s", err.Error())
		// Fallback to default title and message
		title = "Network Configuration Change"
		message = "Network configuration has been modified"
	}

	return &ChangeItem{
		ChangeID:             params.ChangeID,
		ChangeTimestampMicro: currentTime,
		Title:                title,
		Message:              message,
	}
}

// Helper functions

// compareStringSlices compares two string slices for equality.
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
