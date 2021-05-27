package smart

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	defSmartCtlPath = "/usr/bin/smartctl"
	defNvmePath     = "/usr/bin/nvme"
	inputName       = "smart"
	SampleConfig    = `
[[inputs.smart]]
	## The path to the smartctl executable
  # path_smartctl = "/usr/bin/smartctl"

  ## The path to the nvme-cli executable
  # path_nvme = "/usr/bin/nvme"

	## Gathering interval
	# interval = "10s"

  ## Timeout for the cli command to complete.
  # timeout = "30s"

  ## Optionally specify if vendor specific attributes should be propagated for NVMe disk case
  ## ["auto-on"] - automatically find and enable additional vendor specific disk info
  ## ["vendor1", "vendor2", ...] - e.g. "Intel" enable additional Intel specific disk info
  # enable_extensions = ["auto-on"]

  ## On most platforms used cli utilities requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl or nvme-cli.
  ## Sudo must be configured to allow the telegraf user to run smartctl or nvme-cli
  ## without a password.
  # use_sudo = false

  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stopped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your disks.
  # nocheck = "standby"

  ## Gather all returned S.M.A.R.T. attribute metrics and the detailed
  ## information from each drive into the 'smart_attribute' measurement.
  # attributes = false

  ## Optionally specify devices to exclude from reporting if disks auto-discovery is performed.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset
  ## a scan (smartctl --scan and smartctl --scan -d nvme) for S.M.A.R.T. devices will be done
  ## and all found will be included except for the excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam", "/dev/nvme0"]

	## Customer tags, if set will be seen with every metric.
	[inputs.smart.tags]
		# "key1" = "value1"
		# "key2" = "value2"
`
	l = logger.SLogger(inputName)
)

type nvmeDevice struct {
	name         string
	vendorID     string
	model        string
	serialNumber string
}

type Input struct {
	SmartCtlPath string
	NvmePath     string
	Interval     datakit.Duration
	Timeout      datakit.Duration
	Tags         map[string]string
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return SampleConfig
}

func (*Input) AvailabelArch() []string {
	return datakit.AllArch
}

func (i *Input) Run() {
	l.Info("smartctl input started")

	tick := time.NewTicker(i.Interval.Duration)
	for {
		select {
		case <-tick.C:
			if err := i.gather(); err != nil {
				l.Error(err.Error())
				io.FeedLastError(inputName, err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("smart input exits")

			return
		}
	}
}

// Init performs one time setup of the plugin and returns an error if the configuration is invalid.
func (m *Smart) Init() error {
	//if deprecated `path` (to smartctl binary) is provided in config and `path_smartctl` override does not exist
	if len(m.Path) > 0 && len(m.PathSmartctl) == 0 {
		m.PathSmartctl = m.Path
	}

	//if `path_smartctl` is not provided in config, try to find smartctl binary in PATH
	if len(m.PathSmartctl) == 0 {
		m.PathSmartctl, _ = exec.LookPath("smartctl")
	}

	//if `path_nvme` is not provided in config, try to find nvme binary in PATH
	if len(m.PathNVMe) == 0 {
		m.PathNVMe, _ = exec.LookPath("nvme")
	}

	err := validatePath(m.PathSmartctl)
	if err != nil {
		m.PathSmartctl = ""
		//without smartctl, plugin will not be able to gather basic metrics
		return fmt.Errorf("smartctl not found: verify that smartctl is installed and it is in your PATH (or specified in config): %s", err.Error())
	}

	err = validatePath(m.PathNVMe)
	if err != nil {
		m.PathNVMe = ""
		//without nvme, plugin will not be able to gather vendor specific attributes (but it can work without it)
		m.Log.Warnf("nvme not found: verify that nvme is installed and it is in your PATH (or specified in config) to gather vendor specific attributes: %s", err.Error())
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the SMART tools gather.
func (m *Input) gather(acc telegraf.Accumulator) error {
	var err error
	var scannedNVMeDevices []string
	var scannedNonNVMeDevices []string

	devicesFromConfig := m.Devices
	isNVMe := len(m.PathNVMe) != 0
	isVendorExtension := len(m.EnableExtensions) != 0

	if len(m.Devices) != 0 {
		m.getAttributes(acc, devicesFromConfig)

		// if nvme-cli is present, vendor specific attributes can be gathered
		if isVendorExtension && isNVMe {
			scannedNVMeDevices, _, err = m.scanAllDevices(true)
			if err != nil {
				return err
			}
			NVMeDevices := distinguishNVMeDevices(devicesFromConfig, scannedNVMeDevices)

			m.getVendorNVMeAttributes(acc, NVMeDevices)
		}
		return nil
	}
	scannedNVMeDevices, scannedNonNVMeDevices, err = m.scanAllDevices(false)
	if err != nil {
		return err
	}
	var devicesFromScan []string
	devicesFromScan = append(devicesFromScan, scannedNVMeDevices...)
	devicesFromScan = append(devicesFromScan, scannedNonNVMeDevices...)

	m.getAttributes(acc, devicesFromScan)
	if isVendorExtension && isNVMe {
		m.getVendorNVMeAttributes(acc, scannedNVMeDevices)
	}
	return nil
}

func (m *Smart) scanAllDevices(ignoreExcludes bool) ([]string, []string, error) {
	// this will return all devices (including NVMe devices) for smartctl version >= 7.0
	// for older versions this will return non NVMe devices
	devices, err := m.scanDevices(ignoreExcludes, "--scan")
	if err != nil {
		return nil, nil, err
	}

	// this will return only NVMe devices
	NVMeDevices, err := m.scanDevices(ignoreExcludes, "--scan", "--device=nvme")
	if err != nil {
		return nil, nil, err
	}

	// to handle all versions of smartctl this will return only non NVMe devices
	nonNVMeDevices := difference(devices, NVMeDevices)
	return NVMeDevices, nonNVMeDevices, nil
}

// Scan for S.M.A.R.T. devices from smartctl
func (m *Smart) scanDevices(ignoreExcludes bool, scanArgs ...string) ([]string, error) {
	out, err := runCmd(m.Timeout, m.UseSudo, m.PathSmartctl, scanArgs...)
	if err != nil {
		return []string{}, fmt.Errorf("failed to run command '%s %s': %s - %s", m.PathSmartctl, scanArgs, err, string(out))
	}
	var devices []string
	for _, line := range strings.Split(string(out), "\n") {
		dev := strings.Split(line, " ")
		if len(dev) <= 1 {
			continue
		}
		if !ignoreExcludes {
			if !excludedDev(m.Excludes, strings.TrimSpace(dev[0])) {
				devices = append(devices, strings.TrimSpace(dev[0]))
			}
		} else {
			devices = append(devices, strings.TrimSpace(dev[0]))
		}
	}
	return devices, nil
}

func (i *Input) getAttributes(acc telegraf.Accumulator, devices []string) {
	var wg sync.WaitGroup
	wg.Add(len(devices))

	for _, device := range devices {
		go gatherDisk(acc, m.Timeout, m.UseSudo, m.Attributes, m.PathSmartctl, m.Nocheck, device, &wg)
	}

	wg.Wait()
}

func (m *Smart) getVendorNVMeAttributes(acc telegraf.Accumulator, devices []string) {
	NVMeDevices := getDeviceInfoForNVMeDisks(acc, devices, m.PathNVMe, m.Timeout, m.UseSudo)

	var wg sync.WaitGroup

	for _, device := range NVMeDevices {
		if contains(m.EnableExtensions, "auto-on") {
			switch device.vendorID {
			case intelVID:
				wg.Add(1)
				go gatherIntelNVMeDisk(acc, m.Timeout, m.UseSudo, m.PathNVMe, device, &wg)
			}
		} else if contains(m.EnableExtensions, "Intel") && device.vendorID == intelVID {
			wg.Add(1)
			go gatherIntelNVMeDisk(acc, m.Timeout, m.UseSudo, m.PathNVMe, device, &wg)
		}
	}
	wg.Wait()
}

func distinguishNVMeDevices(userDevices []string, availableNVMeDevices []string) []string {
	var NVMeDevices []string

	for _, userDevice := range userDevices {
		for _, NVMeDevice := range availableNVMeDevices {
			// double check. E.g. in case when nvme0 is equal nvme0n1, will check if "nvme0" part is present.
			if strings.Contains(NVMeDevice, userDevice) || strings.Contains(userDevice, NVMeDevice) {
				NVMeDevices = append(NVMeDevices, userDevice)
			}
		}
	}
	return NVMeDevices
}

func excludedDev(excludes []string, deviceLine string) bool {
	device := strings.Split(deviceLine, " ")
	if len(device) != 0 {
		for _, exclude := range excludes {
			if device[0] == exclude {
				return true
			}
		}
	}
	return false
}

func getDeviceInfoForNVMeDisks(acc telegraf.Accumulator, devices []string, nvme string, timeout config.Duration, useSudo bool) []nvmeDevice {
	var NVMeDevices []nvmeDevice

	for _, device := range devices {
		vid, sn, mn, err := gatherNVMeDeviceInfo(nvme, device, timeout, useSudo)
		if err != nil {
			acc.AddError(fmt.Errorf("cannot find device info for %s device", device))
			continue
		}
		newDevice := nvmeDevice{
			name:         device,
			vendorID:     vid,
			model:        mn,
			serialNumber: sn,
		}
		NVMeDevices = append(NVMeDevices, newDevice)
	}
	return NVMeDevices
}

func gatherNVMeDeviceInfo(nvme, device string, timeout config.Duration, useSudo bool) (string, string, string, error) {
	args := []string{"id-ctrl"}
	args = append(args, strings.Split(device, " ")...)
	out, err := runCmd(timeout, useSudo, nvme, args...)
	if err != nil {
		return "", "", "", err
	}
	outStr := string(out)

	vid, sn, mn, err := findNVMeDeviceInfo(outStr)

	return vid, sn, mn, err
}

func findNVMeDeviceInfo(output string) (string, string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var vid, sn, mn string

	for scanner.Scan() {
		line := scanner.Text()

		if matches := nvmeIDCtrlExpressionPattern.FindStringSubmatch(line); len(matches) > 2 {
			matches[1] = strings.TrimSpace(matches[1])
			matches[2] = strings.TrimSpace(matches[2])
			if matches[1] == "vid" {
				if _, err := fmt.Sscanf(matches[2], "%s", &vid); err != nil {
					return "", "", "", err
				}
			}
			if matches[1] == "sn" {
				sn = matches[2]
			}
			if matches[1] == "mn" {
				mn = matches[2]
			}
		}
	}
	return vid, sn, mn, nil
}

func gatherIntelNVMeDisk(acc telegraf.Accumulator, timeout config.Duration, usesudo bool, nvme string, device nvmeDevice, wg *sync.WaitGroup) {
	defer wg.Done()

	args := []string{"intel", "smart-log-add"}
	args = append(args, strings.Split(device.name, " ")...)
	out, e := runCmd(timeout, usesudo, nvme, args...)
	outStr := string(out)

	_, er := exitStatus(e)
	if er != nil {
		acc.AddError(fmt.Errorf("failed to run command '%s %s': %s - %s", nvme, strings.Join(args, " "), e, outStr))
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(outStr))

	for scanner.Scan() {
		line := scanner.Text()
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["device"] = path.Base(device.name)
		tags["model"] = device.model
		tags["serial_no"] = device.serialNumber

		if matches := intelExpressionPattern.FindStringSubmatch(line); len(matches) > 3 {
			matches[1] = strings.TrimSpace(matches[1])
			matches[3] = strings.TrimSpace(matches[3])
			if attr, ok := intelAttributes[matches[1]]; ok {
				tags["name"] = attr.Name
				if attr.ID != "" {
					tags["id"] = attr.ID
				}

				parse := parseCommaSeparatedIntWithAccumulator
				if attr.Parse != nil {
					parse = attr.Parse
				}

				if err := parse(acc, fields, tags, matches[3]); err != nil {
					continue
				}
			}
		}
	}
}

func gatherDisk(acc telegraf.Accumulator, timeout config.Duration, usesudo, collectAttributes bool, smartctl, nocheck, device string, wg *sync.WaitGroup) {
	defer wg.Done()
	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n
	args := []string{"--info", "--health", "--attributes", "--tolerance=verypermissive", "-n", nocheck, "--format=brief"}
	args = append(args, strings.Split(device, " ")...)
	out, e := runCmd(timeout, usesudo, smartctl, args...)
	outStr := string(out)

	// Ignore all exit statuses except if it is a command line parse error
	exitStatus, er := exitStatus(e)
	if er != nil {
		acc.AddError(fmt.Errorf("failed to run command '%s %s': %s - %s", smartctl, strings.Join(args, " "), e, outStr))
		return
	}

	deviceTags := map[string]string{}
	deviceNode := strings.Split(device, " ")[0]
	deviceTags["device"] = path.Base(deviceNode)
	deviceFields := make(map[string]interface{})
	deviceFields["exit_status"] = exitStatus

	scanner := bufio.NewScanner(strings.NewReader(outStr))

	for scanner.Scan() {
		line := scanner.Text()

		model := modelInfo.FindStringSubmatch(line)
		if len(model) > 2 {
			deviceTags["model"] = model[2]
		}

		serial := serialInfo.FindStringSubmatch(line)
		if len(serial) > 1 {
			deviceTags["serial_no"] = serial[1]
		}

		wwn := wwnInfo.FindStringSubmatch(line)
		if len(wwn) > 1 {
			deviceTags["wwn"] = strings.Replace(wwn[1], " ", "", -1)
		}

		capacity := userCapacityInfo.FindStringSubmatch(line)
		if len(capacity) > 1 {
			deviceTags["capacity"] = strings.Replace(capacity[1], ",", "", -1)
		}

		enabled := smartEnabledInfo.FindStringSubmatch(line)
		if len(enabled) > 1 {
			deviceTags["enabled"] = enabled[1]
		}

		health := smartOverallHealth.FindStringSubmatch(line)
		if len(health) > 2 {
			deviceFields["health_ok"] = health[2] == "PASSED" || health[2] == "OK"
		}

		tags := map[string]string{}
		fields := make(map[string]interface{})

		if collectAttributes {
			keys := [...]string{"device", "model", "serial_no", "wwn", "capacity", "enabled"}
			for _, key := range keys {
				if value, ok := deviceTags[key]; ok {
					tags[key] = value
				}
			}
		}

		attr := attribute.FindStringSubmatch(line)
		if len(attr) > 1 {
			// attribute has been found, add it only if collectAttributes is true
			if collectAttributes {
				tags["id"] = attr[1]
				tags["name"] = attr[2]
				tags["flags"] = attr[3]

				fields["exit_status"] = exitStatus
				if i, err := strconv.ParseInt(attr[4], 10, 64); err == nil {
					fields["value"] = i
				}
				if i, err := strconv.ParseInt(attr[5], 10, 64); err == nil {
					fields["worst"] = i
				}
				if i, err := strconv.ParseInt(attr[6], 10, 64); err == nil {
					fields["threshold"] = i
				}

				tags["fail"] = attr[7]
				if val, err := parseRawValue(attr[8]); err == nil {
					fields["raw_value"] = val
				}

				acc.AddFields("smart_attribute", fields, tags)
			}

			// If the attribute matches on the one in deviceFieldIds
			// save the raw value to a field.
			if field, ok := deviceFieldIds[attr[1]]; ok {
				if val, err := parseRawValue(attr[8]); err == nil {
					deviceFields[field] = val
				}
			}
		} else {
			// what was found is not a vendor attribute
			if matches := sasNvmeAttr.FindStringSubmatch(line); len(matches) > 2 {
				if attr, ok := sasNvmeAttributes[matches[1]]; ok {
					tags["name"] = attr.Name
					if attr.ID != "" {
						tags["id"] = attr.ID
					}

					parse := parseCommaSeparatedInt
					if attr.Parse != nil {
						parse = attr.Parse
					}

					if err := parse(fields, deviceFields, matches[2]); err != nil {
						continue
					}
					// if the field is classified as an attribute, only add it
					// if collectAttributes is true
					if collectAttributes {
						acc.AddFields("smart_attribute", fields, tags)
					}
				}
			}
		}
	}
	acc.AddFields("smart_device", deviceFields, deviceTags)
}

// Command line parse errors are denoted by the exit code having the 0 bit set.
// All other errors are drive/communication errors and should be ignored.
func exitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}
	return 0, err
}

func contains(args []string, element string) bool {
	for _, arg := range args {
		if arg == element {
			return true
		}
	}
	return false
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func validatePath(path string) error {
	pathInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("provided path does not exist: [%s]", path)
	}
	if mode := pathInfo.Mode(); !mode.IsRegular() {
		return fmt.Errorf("provided path does not point to a regular file: [%s]", path)
	}

	return nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
