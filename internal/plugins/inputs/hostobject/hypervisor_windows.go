// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package hostobject

import (
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

// isVirtual detects if the current Windows system is running in a virtual machine.
// It checks in the following order:
// 1. WMI Win32_ComputerSystem Manufacturer/Model
// 2. WMI Win32_BaseBoard Manufacturer
// 3. Registry HKLM\System\CurrentControlSet\Services
// 4. Processor ID signature
func isVirtual() bool {
	// Check 1: WMI Win32_ComputerSystem
	if isWMIVirtualMachine() {
		return true
	}

	// Check 2: WMI Win32_BaseBoard
	if isWMIBaseBoardVM() {
		return true
	}

	// Check 3: Registry check
	if isRegistryVM() {
		return true
	}

	// Check 4: Processor signature
	if isCPUSignatureVM() {
		return true
	}

	return false
}

// getHypervisorType returns the type of hypervisor on Windows.
func getHypervisorType() string {
	// Try WMI first
	if hypervisor := getWMIHypervisor(); hypervisor != "" {
		return hypervisor
	}

	// Try registry
	if hypervisor := getRegistryHypervisor(); hypervisor != "" {
		return hypervisor
	}

	// Try CPU signature
	if hypervisor := getCPUSignatureHypervisor(); hypervisor != "" {
		return hypervisor
	}

	return ""
}

// isWMIVirtualMachine checks WMI for VM indicators via Win32_ComputerSystem
func isWMIVirtualMachine() bool {
	manufacturer := getWMIManufacturer()
	model := getWMIModel()

	vmIndicators := map[string]bool{
		"vmware":     true,
		"virtualbox": true,
		"innotek":    true,
		"parallels":  true,
		"xen":        true,
		"qemu":       true,
		"kvm":        true,
		"bochs":      true,
		"hyper-v":    true,
		"microsoft":  true,
		"citrix":     true,
		"vm":         true,
		"virtual":    true,
	}

	manufacturer = strings.ToLower(manufacturer)
	model = strings.ToLower(model)

	for indicator := range vmIndicators {
		if strings.Contains(manufacturer, indicator) || strings.Contains(model, indicator) {
			return true
		}
	}

	return false
}

// isWMIBaseBoardVM checks WMI Win32_BaseBoard manufacturer for VM indicators
func isWMIBaseBoardVM() bool {
	baseBoardMfg := getWMIBaseBoardManufacturer()
	baseBoardMfg = strings.ToLower(baseBoardMfg)

	vmIndicators := []string{
		"vmware", "virtualbox", "parallels", "xen", "qemu", "kvm", "bochs", "hyper-v",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(baseBoardMfg, indicator) {
			return true
		}
	}

	return false
}

// isRegistryVM checks Windows registry for VM-related services
func isRegistryVM() bool {
	vmServices := []string{
		"VMware",
		"VirtualBox",
		"Hyper-V",
		"Parallels",
		"KVM",
		"Xen",
		"QEMU",
		"Bochs",
	}

	for _, service := range vmServices {
		regPath := `System\CurrentControlSet\Services\` + service
		var key windows.Handle
		err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, windows.StringToUTF16Ptr(regPath), 0, windows.KEY_READ, &key)
		if err == nil {
			windows.RegCloseKey(key)
			return true
		}
	}

	return false
}

// isCPUSignatureVM checks CPU signature for hypervisor flag using WMI
func isCPUSignatureVM() bool {
	// Check if Hyper-V or other hypervisor is present via WMI
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject Win32_Processor | Select-Object -ExpandProperty Manufacturer")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	manufacturer := strings.TrimSpace(string(output))
	manufacturer = strings.ToLower(manufacturer)

	// Check for hypervisor signatures
	vmIndicators := []string{
		"vmware", "virtualbox", "microsoft", "xen", "qemu",
		"parallels", "kvm", "bochs",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(manufacturer, indicator) {
			return true
		}
	}

	return false
}

// getWMIManufacturer gets Win32_ComputerSystem Manufacturer via WMI using PowerShell
func getWMIManufacturer() string {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject Win32_ComputerSystem | Select-Object -ExpandProperty Manufacturer")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getWMIModel gets Win32_ComputerSystem Model via WMI using PowerShell
func getWMIModel() string {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject Win32_ComputerSystem | Select-Object -ExpandProperty Model")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getWMIBaseBoardManufacturer gets Win32_BaseBoard Manufacturer via WMI using PowerShell
func getWMIBaseBoardManufacturer() string {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject Win32_BaseBoard | Select-Object -ExpandProperty Manufacturer")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getWMIHypervisor extracts hypervisor type from WMI
func getWMIHypervisor() string {
	manufacturer := getWMIManufacturer()
	model := getWMIModel()
	baseboard := getWMIBaseBoardManufacturer()

	manufacturer = strings.ToLower(manufacturer)
	model = strings.ToLower(model)
	baseboard = strings.ToLower(baseboard)

	if strings.Contains(manufacturer, "vmware") || strings.Contains(model, "vmware") {
		return "vmware"
	}
	if strings.Contains(manufacturer, "virtualbox") || strings.Contains(baseboard, "virtualbox") {
		return "virtualbox"
	}
	if strings.Contains(manufacturer, "parallels") || strings.Contains(baseboard, "parallels") {
		return "parallels"
	}
	if strings.Contains(manufacturer, "microsoft") || strings.Contains(model, "hyper-v") {
		return "hyperv"
	}
	if strings.Contains(manufacturer, "xen") || strings.Contains(model, "xen") {
		return "xen"
	}
	if strings.Contains(manufacturer, "qemu") || strings.Contains(model, "qemu") {
		return "qemu"
	}
	if strings.Contains(manufacturer, "kvm") || strings.Contains(model, "kvm") {
		return "kvm"
	}

	return ""
}

// getRegistryHypervisor extracts hypervisor type from registry
func getRegistryHypervisor() string {
	vmServiceMap := map[string]string{
		"VMware":     "vmware",
		"VirtualBox": "virtualbox",
		"Hyper-V":    "hyperv",
		"Parallels":  "parallels",
		"KVM":        "kvm",
		"Xen":        "xen",
		"QEMU":       "qemu",
		"Bochs":      "bochs",
	}

	for service, hypervisorType := range vmServiceMap {
		regPath := `System\CurrentControlSet\Services\` + service
		var key windows.Handle
		err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, windows.StringToUTF16Ptr(regPath), 0, windows.KEY_READ, &key)
		if err == nil {
			windows.RegCloseKey(key)
			return hypervisorType
		}
	}

	return ""
}

// getCPUSignatureHypervisor extracts hypervisor type from CPU signature using WMI
func getCPUSignatureHypervisor() string {
	// Check CPU Manufacturer for hypervisor signatures
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject Win32_Processor | Select-Object -ExpandProperty Manufacturer")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	manufacturer := strings.TrimSpace(strings.ToLower(string(output)))

	if strings.Contains(manufacturer, "vmware") {
		return "vmware"
	}
	if strings.Contains(manufacturer, "virtualbox") {
		return "virtualbox"
	}
	if strings.Contains(manufacturer, "microsoft") {
		return "hyperv"
	}
	if strings.Contains(manufacturer, "xen") {
		return "xen"
	}
	if strings.Contains(manufacturer, "qemu") {
		return "qemu"
	}
	if strings.Contains(manufacturer, "parallels") {
		return "parallels"
	}
	if strings.Contains(manufacturer, "kvm") {
		return "kvm"
	}
	if strings.Contains(manufacturer, "bochs") {
		return "bochs"
	}

	return ""
}
