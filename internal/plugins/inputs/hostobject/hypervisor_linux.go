// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package hostobject

import (
	"bufio"
	"os"
	"strings"
)

// isVirtual detects if the current Linux system is running in a virtual machine.
// It checks in the following order:
// 1. /sys/hypervisor/ directory (Xen)
// 2. /proc/cpuinfo flags (KVM, VMware, etc.)
// 3. DMI System Product Name
// 4. /proc/vz/ (OpenVZ)
// 5. /run/systemd/container (Container environment).
func isVirtual() bool {
	// Check 1: Xen hypervisor
	if isXenVM() {
		return true
	}

	// Check 2: Check /proc/cpuinfo for VM flags
	if isCPUFlagsVM() {
		return true
	}

	// Check 3: Check DMI System Product Name
	if isDMIVM() {
		return true
	}

	// Check 4: OpenVZ
	if isOpenVZVM() {
		return true
	}

	// Check 5: Container environment
	if isContainerEnv() {
		return true
	}

	return false
}

// getHypervisorType returns the type of hypervisor on Linux.
func getHypervisorType() string {
	// Check Xen
	if _, err := os.Stat("/sys/hypervisor/type"); err == nil {
		if data, err := os.ReadFile("/sys/hypervisor/type"); err == nil {
			return strings.TrimSpace(string(data))
		}
		return "xen"
	}

	// Check /proc/cpuinfo for VM flags
	if hypervisor := getCPUHypervisor(); hypervisor != "" {
		return hypervisor
	}

	// Check DMI
	if dmiType := getDMIHypervisor(); dmiType != "" {
		return dmiType
	}

	// Check OpenVZ
	if _, err := os.Stat("/proc/vz"); err == nil {
		return "openvz"
	}

	// Check container
	if containerType := getContainerType(); containerType != "" {
		return containerType
	}

	return ""
}

// isXenVM checks if running on Xen hypervisor.
func isXenVM() bool {
	_, err := os.Stat("/sys/hypervisor")
	return err == nil
}

// isCPUFlagsVM checks /proc/cpuinfo for VM-related CPU flags.
func isCPUFlagsVM() bool {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "flags") {
			// Check for common hypervisor flags
			if strings.Contains(line, "hypervisor") {
				return true
			}
		}
	}
	return false
}

// getCPUHypervisor extracts hypervisor type from /proc/cpuinfo.
func getCPUHypervisor() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			// Check for specific hypervisor signatures in model name
			if strings.Contains(line, "KVM") {
				return "kvm"
			}
			if strings.Contains(line, "QEMU") {
				return "qemu"
			}
		}
	}
	return ""
}

// isDMIVM checks DMI system product name for VM indicators.
func isDMIVM() bool {
	dmiPath := "/sys/devices/virtual/dmi/id/system_product_name"
	data, err := os.ReadFile(dmiPath)
	if err != nil {
		return false
	}

	product := strings.TrimSpace(strings.ToLower(string(data)))
	vmIndicators := []string{
		"vmware", "virtualbox", "xen", "hyperv", "qemu",
		"kvm", "bochs", "parallels", "vbox",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(product, indicator) {
			return true
		}
	}

	return false
}

// getDMIHypervisor extracts specific hypervisor type from DMI.
func getDMIHypervisor() string {
	dmiPath := "/sys/devices/virtual/dmi/id/system_product_name"
	data, err := os.ReadFile(dmiPath)
	if err != nil {
		return ""
	}

	product := strings.TrimSpace(strings.ToLower(string(data)))

	// Check for specific hypervisor types
	if strings.Contains(product, "vmware") {
		return "vmware"
	}
	if strings.Contains(product, "virtualbox") || strings.Contains(product, "vbox") {
		return "virtualbox"
	}
	if strings.Contains(product, "xen") {
		return "xen"
	}
	if strings.Contains(product, "hyperv") {
		return "hyperv"
	}
	if strings.Contains(product, "qemu") {
		return "qemu"
	}
	if strings.Contains(product, "kvm") {
		return "kvm"
	}
	if strings.Contains(product, "bochs") {
		return "bochs"
	}
	if strings.Contains(product, "parallels") {
		return "parallels"
	}

	return ""
}

// isOpenVZVM checks for OpenVZ environment.
func isOpenVZVM() bool {
	_, err := os.Stat("/proc/vz")
	return err == nil
}

// isContainerEnv checks if running in a container (Docker, Kubernetes, etc.)
func isContainerEnv() bool {
	_, err := os.Stat("/.dockerenv")
	if err == nil {
		return true
	}

	_, err = os.Stat("/run/systemd/container")
	if err == nil {
		return true
	}

	// Check cgroup for container signature
	cgroup, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		cgroupStr := string(cgroup)
		if strings.Contains(cgroupStr, "docker") || strings.Contains(cgroupStr, "lxc") {
			return true
		}
	}

	return false
}

// getContainerType identifies the container type.
func getContainerType() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	if data, err := os.ReadFile("/run/systemd/container"); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Check cgroup
	cgroup, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		cgroupStr := string(cgroup)
		if strings.Contains(cgroupStr, "docker") {
			return "docker"
		}
		if strings.Contains(cgroupStr, "lxc") {
			return "lxc"
		}
	}

	return ""
}
