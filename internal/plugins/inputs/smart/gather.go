// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package smart

import (
	"bufio"
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/command"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/strarr"
)

// Gather takes in an accumulator and adds the metrics that the SMART tools gather.
func (ipt *Input) gather() error {
	var (
		err                   error
		scannedNVMeDevices    []string
		scannedNonNVMeDevices []string
		isNVMe                = len(ipt.NvmePath) != 0
		isVendorExtension     = len(ipt.EnableExtensions) != 0
	)

	if len(ipt.Devices) != 0 {
		if err := ipt.getAttributes(ipt.Devices); err != nil {
			return err
		}

		// if nvme-cli is present, vendor specific attributes can be gathered
		if isVendorExtension && isNVMe {
			if scannedNVMeDevices, _, err = ipt.scanAllDevices(true); err != nil {
				return err
			}
			if err = ipt.getVendorNVMeAttributes(distinguishNVMeDevices(ipt.Devices, scannedNVMeDevices)); err != nil {
				return err
			}
		}
	} else {
		if scannedNVMeDevices, scannedNonNVMeDevices, err = ipt.scanAllDevices(false); err != nil {
			return err
		}

		if err := ipt.getAttributes(append(scannedNVMeDevices, scannedNonNVMeDevices...)); err != nil {
			return err
		}

		if isVendorExtension && isNVMe {
			return ipt.getVendorNVMeAttributes(scannedNVMeDevices)
		}
	}

	return nil
}

// Get info and attributes for each S.M.A.R.T. device.
func (ipt *Input) getAttributes(devices []string) error {
	start := time.Now()

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_smart"})
	for _, device := range devices {
		func(device string) {
			g.Go(func(ctx context.Context) error {
				if pt, err := ipt.gatherDisk(device); err != nil {
					l.Errorf("gatherDisk: %s", err.Error())

					metrics.FeedLastError(inputName, err.Error())
				} else {
					return ipt.feeder.Feed(point.Metric, []*point.Point{pt},
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithSource(inputName),
					)
				}

				return nil
			})
		}(device)
	}

	return g.Wait()
}

func (ipt *Input) getVendorNVMeAttributes(devices []string) error {
	start := time.Now()
	nvmeDevices := getDeviceInfoForNVMeDisks(devices, ipt.NvmePath, ipt.Timeout.Duration, ipt.UseSudo)

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_smart"})
	for _, device := range nvmeDevices {
		if strarr.Contains(ipt.EnableExtensions, "auto-on") {
			if device.vendorID == intelVID {
				func(device nvmeDevice) {
					g.Go(func(ctx context.Context) error {
						if pt, err := ipt.gatherIntelNVMeDisk(device); err != nil {
							l.Errorf("gatherIntelNVMeDisk: %s", err.Error())

							metrics.FeedLastError(inputName, err.Error())
						} else {
							return ipt.feeder.Feed(point.Metric, []*point.Point{pt},
								dkio.WithCollectCost(time.Since(start)),
								dkio.WithSource(inputName),
							)
						}
						return nil
					})
				}(device)
			}
		} else if strarr.Contains(ipt.EnableExtensions, "Intel") && device.vendorID == intelVID {
			func(device nvmeDevice) {
				g.Go(func(ctx context.Context) error {
					if pt, err := ipt.gatherIntelNVMeDisk(device); err != nil {
						l.Errorf("gatherIntelNVMeDisk: %s", err.Error())
						metrics.FeedLastError(inputName, err.Error())
					} else {
						return ipt.feeder.Feed(point.Metric, []*point.Point{pt},
							dkio.WithCollectCost(time.Since(start)),
							dkio.WithSource(inputName),
						)
					}

					return nil
				})
			}(device)
		}
	}

	return g.Wait()
}

func distinguishNVMeDevices(userDevices []string, availableNVMeDevices []string) []string {
	var nvmeDevices []string
	for _, userDevice := range userDevices {
		for _, NVMeDevice := range availableNVMeDevices {
			// double check. E.g. in case when nvme0 is equal nvme0n1, will check if "nvme0" part is present.
			if strings.Contains(NVMeDevice, userDevice) || strings.Contains(userDevice, NVMeDevice) {
				nvmeDevices = append(nvmeDevices, userDevice)
			}
		}
	}

	return nvmeDevices
}

// Scan for S.M.A.R.T. devices from smartctl.
func (ipt *Input) scanDevices(ignoreExcludes bool, scanArgs ...string) ([]string, error) {
	l.Debugf("run command %s %v", ipt.SmartCtlPath, scanArgs)
	output, err := command.RunWithTimeout(ipt.Timeout.Duration, ipt.UseSudo, ipt.SmartCtlPath, scanArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%s %s': %w - %s", ipt.SmartCtlPath, scanArgs, err, string(output))
	}

	var devices []string
	for _, line := range strings.Split(string(output), "\n") {
		dev := strings.Split(line, " ")
		if len(dev) <= 1 {
			continue
		}
		if !ignoreExcludes {
			if !excludedDevice(ipt.Excludes, strings.TrimSpace(dev[0])) {
				devices = append(devices, strings.TrimSpace(dev[0]))
			}
		} else {
			devices = append(devices, strings.TrimSpace(dev[0]))
		}
	}

	return devices, nil
}

func (ipt *Input) scanAllDevices(ignoreExcludes bool) ([]string, []string, error) {
	// this will return all devices (including NVMe devices) for smartctl version >= 7.0
	// for older versions this will return non NVMe devices
	devices, err := ipt.scanDevices(ignoreExcludes, "--scan")
	if err != nil {
		return nil, nil, err
	}

	// this will return only NVMe devices
	nvmeDevices, err := ipt.scanDevices(ignoreExcludes, "--scan", "--device=nvme")
	if err != nil {
		return nil, nil, err
	}

	// to handle all versions of smartctl this will return only non NVMe devices
	nonNVMeDevices := strarr.Differ(devices, nvmeDevices)

	return nvmeDevices, nonNVMeDevices, nil
}

func excludedDevice(excludes []string, deviceLine string) bool {
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

func gatherNVMeDeviceInfo(nvme, device string, timeout time.Duration, useSudo bool) (string, string, string, error) {
	args := append([]string{"id-ctrl"}, strings.Split(device, " ")...)
	l.Debugf("run command %s %v", nvme, args)
	output, err := command.RunWithTimeout(timeout, useSudo, nvme, args...)
	if err != nil {
		return "", "", "", err
	}

	return findNVMeDeviceInfo(string(output))
}

func getDeviceInfoForNVMeDisks(devices []string, nvme string, timeout time.Duration, useSudo bool) []nvmeDevice {
	var nvmeDevices []nvmeDevice
	for _, device := range devices {
		vid, sn, mn, err := gatherNVMeDeviceInfo(nvme, device, timeout, useSudo)
		if err != nil {
			l.Errorf("gatherNVMeDeviceInfo: %s", err)

			metrics.FeedLastError(inputName, fmt.Sprintf("cannot find device info for %s device", device))
			continue
		}
		newDevice := nvmeDevice{
			name:         device,
			vendorID:     vid,
			model:        mn,
			serialNumber: sn,
		}
		nvmeDevices = append(nvmeDevices, newDevice)
	}

	return nvmeDevices
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

func (ipt *Input) gatherIntelNVMeDisk(device nvmeDevice) (*point.Point, error) {
	var (
		args = append([]string{"intel", "smart-log-add"}, strings.Split(device.name, " ")...)
		kvs  = point.NewTags(ipt.mergedTags)
	)

	l.Debugf("run command %s %v", ipt.NvmePath, args)
	output, err := command.RunWithTimeout(ipt.Timeout.Duration, ipt.UseSudo, ipt.NvmePath, args...)
	if _, err = command.ExitStatus(err); err != nil {
		return nil, fmt.Errorf("failed to run command '%s %s': %w - %s",
			ipt.NvmePath, strings.Join(args, " "), err, string(output))
	}

	kvs = kvs.AddTag("device", path.Base(device.name)).
		AddTag("model", device.model).
		AddTag("serial_no", device.serialNumber)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := intelExpressionPattern.FindStringSubmatch(line); len(matches) > 3 {
			matches[1] = strings.TrimSpace(matches[1])
			matches[3] = strings.TrimSpace(matches[3])
			if attr, ok := intelAttributes[matches[1]]; ok {
				parse := parseCommaSeparatedIntWithCache
				if attr.Parse != nil {
					parse = attr.Parse
				}

				if x, err := parse(attr.Name, matches[3], kvs); err != nil {
					continue
				} else {
					kvs = x
				}
			}
		}
	}

	return point.NewPointV2("smart", kvs, append(point.DefaultMetricOptions(), point.WithTime(ipt.ptsTime))...), nil
}

func parseInt(str string) int64 {
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}

	return 0
}

func parseRawValue(rawVal string) (int64, error) {
	// Integer
	if i, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
		return i, nil
	}

	// Duration: 65h+33m+09.259s
	unit := regexp.MustCompile("^(.*)([hms])$")
	parts := strings.Split(rawVal, "+")
	if len(parts) == 0 {
		return 0, fmt.Errorf("couldn't parse RAW_VALUE '%s'", rawVal)
	}

	duration := int64(0)
	for _, part := range parts {
		timePart := unit.FindStringSubmatch(part)
		if len(timePart) == 0 {
			continue
		}
		switch timePart[2] {
		case "h":
			duration += parseInt(timePart[1]) * int64(3600)
		case "m":
			duration += parseInt(timePart[1]) * int64(60)
		case "s":
			// drop fractions of seconds
			duration += parseInt(strings.Split(timePart[1], ".")[0])
		default:
			// Unknown, ignore
		}
	}
	return duration, nil
}

// ipt.getCustomerTags(), ipt.Timeout.Duration, ipt.UseSudo, ipt.SmartCtlPath, ipt.NoCheck,.
func (ipt *Input) gatherDisk(device string) (*point.Point, error) {
	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n
	var (
		kvs point.KVs
	)

	output := ipt.getter.Get(strings.Split(device, " "))
	// Ignore all exit statuses except if it is a command line parse error
	exitStatus := ipt.getter.ExitStatus()

	kvs = point.NewTags(ipt.mergedTags).
		AddTag("device", path.Base(strings.Split(device, " ")[0]))

	if exitStatus == 0 {
		kvs = kvs.AddTag("exit_status", "success")
	} else {
		kvs = kvs.AddTag("exit_status", "failed")
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()

		if arr := modelInfo.FindStringSubmatch(line); len(arr) > 2 {
			kvs = kvs.AddTag("model", arr[2])
		}

		if arr := serialInfo.FindStringSubmatch(line); len(arr) > 1 {
			kvs = kvs.AddTag("serial_no", arr[1])
		}

		if arr := wwnInfo.FindStringSubmatch(line); len(arr) > 1 {
			kvs = kvs.AddTag("wwn", strings.ReplaceAll(arr[1], " ", ""))
		}

		if arr := userCapacityInfo.FindStringSubmatch(line); len(arr) > 1 {
			x := strings.ReplaceAll(arr[1], ",", "") // convert 123,456,789 -> 123456789

			if c, err := strconv.Atoi(x); err == nil {
				kvs = kvs.AddV2("capacity", c, true)
			} else {
				l.Warnf("invalid capacity: %q", x)
			}
		}

		if arr := smartEnabledInfo.FindStringSubmatch(line); len(arr) > 1 {
			kvs = kvs.AddTag("enabled", arr[1])
		}

		if arr := smartOverallHealth.FindStringSubmatch(line); len(arr) > 2 {
			kvs = kvs.AddTag("health_ok", arr[2])
		}

		if arr := attribute.FindStringSubmatch(line); len(arr) > 1 {
			// attribute has been found
			name := strings.ToLower(arr[2])
			kvs = kvs.AddTag("flags", arr[3])

			if i, err := strconv.ParseInt(arr[4], 10, 64); err == nil {
				kvs = kvs.AddV2(name+"_value", i, true)
			}

			if i, err := strconv.ParseInt(arr[5], 10, 64); err == nil {
				kvs = kvs.AddV2(name+"_worst", i, true)
			}

			if i, err := strconv.ParseInt(arr[6], 10, 64); err == nil {
				kvs = kvs.AddV2(name+"_threshold", i, true)
			}

			kvs = kvs.AddV2("fail", !(arr[7] == "-"), true)
			if val, err := parseRawValue(arr[8]); err == nil {
				kvs = kvs.AddV2(name+"_raw_value", val, true)
			}

			// If the attribute matches on the one in deviceFieldIds save the raw value to a field.
			if field, ok := deviceFieldIds[arr[1]]; ok {
				if val, err := parseRawValue(arr[8]); err == nil {
					kvs = kvs.AddV2(field, val, true)
				}
			}
		} else if arr := sasNvmeAttr.FindStringSubmatch(line); len(arr) > 2 {
			// what was found is not a vendor attribute
			if attr, ok := sasNvmeAttributes[arr[1]]; ok {
				parse := parseCommaSeparatedInt
				if attr.Parse != nil {
					parse = attr.Parse
				}

				if x, err := parse(attr.Name, arr[2], kvs); err != nil {
					continue
				} else {
					kvs = x
				}
			}
		}
	}

	return point.NewPointV2("smart", kvs, point.DefaultMetricOptions()...), nil
}
