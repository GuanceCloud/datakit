// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package lsblk

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var (
	ScsiTypeDisk          = 0x00
	ScsiTypeTape          = 0x01
	ScsiTypePrinter       = 0x02
	ScsiTypeProcessor     = 0x03 // HP scanners use this
	ScsiTypeWorm          = 0x04 // Treated as ROM by our system
	ScsiTypeRom           = 0x05
	ScsiTypeScanner       = 0x06
	ScsiTypeMod           = 0x07 // Magneto-optical disk - treated as ScsiTypeDisk
	ScsiTypeMediumChanger = 0x08
	ScsiTypeComm          = 0x09 // Communications device
	ScsiTypeRaid          = 0x0c
	ScsiTypeEnclosure     = 0x0d // Enclosure Services Device
	ScsiTypeRbc           = 0x0e
	ScsiTypeOsd           = 0x11
	ScsiTypeNoLun         = 0x7f

	PARTITION = "partition"
	DISK      = "disk"
)

func blkdevScsiTypeToName(blkType int) string {
	switch blkType {
	case ScsiTypeDisk:
		return "disk"
	case ScsiTypeTape:
		return "tape"
	case ScsiTypePrinter:
		return "printer"
	case ScsiTypeProcessor:
		return "processor"
	case ScsiTypeWorm:
		return "worm"
	case ScsiTypeRom:
		return "rom"
	case ScsiTypeScanner:
		return "scanner"
	case ScsiTypeMod:
		return "mo-disk"
	case ScsiTypeMediumChanger:
		return "changer"
	case ScsiTypeComm:
		return "comm"
	case ScsiTypeRaid:
		return "raid"
	case ScsiTypeEnclosure:
		return "enclosure"
	case ScsiTypeRbc:
		return "rbc"
	case ScsiTypeOsd:
		return "osd"
	case ScsiTypeNoLun:
		return "no-lun"
	default:
		break
	}

	return ""
}

func hostSys(combineWith ...string) string {
	value := os.Getenv("HOST_SYS")
	if value == "" {
		value = "/sys"
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

func hostProc(combineWith ...string) string {
	value := os.Getenv("HOST_PROC")
	if value == "" {
		value = "/proc"
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

func hostDev(combineWith ...string) string {
	value := os.Getenv("HOST_DEV")
	if value == "" {
		value = "/dev"
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

/* Checks for DM prefix in the device name. */
func isDM(name string) bool {
	return strings.HasPrefix(name, "dm-")
}

func makeName(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	return filepath.Join("/dev/", name)
}

func makeDmName(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	return filepath.Join("/dev/mapper/", name)
}

// Unescape escaped octal chars (like space 040, ampersand 046 and backslash 134) to their real value in fstab fields issue#555.
func unescapeFstab(path string) string {
	escaped, err := strconv.Unquote(`"` + path + `"`)
	if err != nil {
		return path
	}
	return escaped
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLinesOffsetN reads contents from a file and splits them by new lines.
//
//	n >= 0: at most n lines
//	n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return []string{""}, err
	}
	defer func() {
		_ = f.Close()
	}()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

// StringsHas checks the target string slice contains src or not.
func StringsHas(target []string, src string) bool {
	for _, t := range target {
		if strings.TrimSpace(t) == src {
			return true
		}
	}
	return false
}

// getFileSystems returns supported filesystems from /proc/filesystems.
func getFileSystems() ([]string, error) {
	filename := hostProc("filesystems")
	lines, err := ReadLines(filename)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "nodev") {
			ret = append(ret, strings.TrimSpace(line))
			continue
		}
		t := strings.Split(line, "\t")
		if len(t) != 2 || t[1] != "zfs" {
			continue
		}
		ret = append(ret, strings.TrimSpace(t[1]))
	}

	return ret, nil
}

func ulPathReadString(basePath string, path string) string {
	fullPath := filepath.Join(basePath, path)
	data, err := os.ReadFile(fullPath) //nolint:gosec
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func ulPathReadInt(basePath string, path string) (int, error) {
	fullPath := filepath.Join(basePath, path)
	data, err := os.ReadFile(fullPath) //nolint:gosec
	if err != nil {
		return -1, err
	}
	i, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return -1, err
	}
	return i, nil
}

func setDetailInfo(device *BlockDeviceStat, devicePath string) {
	if !device.IsMounted {
		// set Name
		device.Name = makeName(filepath.Base(devicePath))
	}

	setType(device, devicePath)

	// set Parent
	if device.Type == PARTITION {
		parts := strings.Split(devicePath, "/")
		if len(parts) >= 2 {
			device.Parents = append(device.Parents, makeName(parts[len(parts)-2]))
		}
	}

	// set Vendor
	if device.Type == "disk" {
		vendor := ulPathReadString(devicePath, "device/vendor")
		device.Vendor = vendor
	}

	// set RQSize
	rqSizeStr := ulPathReadString(devicePath, "queue/nr_requests")
	if rqSize, err := strconv.ParseFloat(rqSizeStr, 64); err == nil {
		device.RQSize = rqSize
	}

	// set serial
	if device.Type != PARTITION {
		device.Serial = ulPathReadString(devicePath, "device/serial")
	}

	// set state
	if device.Type != PARTITION {
		device.State = ulPathReadString(devicePath, "device/state")
	}

	setCommonDetailInfo(device, devicePath)
}

func setDMDetailInfo(device *BlockDeviceStat, devicePath string) {
	device.IsDM = true
	// set Name
	dmName := ulPathReadString(devicePath, "dm/name")
	device.Name = makeDmName(dmName)

	setType(device, devicePath)

	// set Parent
	if entries, err := os.ReadDir(filepath.Join(devicePath, "slaves")); err == nil {
		for _, entry := range entries {
			device.Parents = append(device.Parents, makeName(entry.Name()))
		}
	}

	// set state
	if suspended, err := ulPathReadInt(devicePath, "dm/suspended"); err == nil {
		if suspended == 0 {
			device.State = "running"
		} else {
			device.State = "suspended"
		}
	}

	setCommonDetailInfo(device, devicePath)
}

func setCommonDetailInfo(device *BlockDeviceStat, devicePath string) {
	setOwnerAndGroup(device, devicePath)

	// 设置fsavail、fssize、fsused%等信息
	setFilesystemStats(device)

	// set Size
	sizeStr := ulPathReadString(devicePath, "size")
	if deviceSize, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		device.Size = deviceSize * 512
	}
}

func setFilesystemStats(device *BlockDeviceStat) {
	if device.MountPoint == "" {
		return
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(device.MountPoint, &stat); err != nil {
		return
	}

	device.FSSize = float64(stat.Blocks * uint64(stat.Bsize))
	device.FSUsed = float64((stat.Blocks - stat.Bfree) * uint64(stat.Bsize))
	device.FSAvail = float64(stat.Bavail * uint64(stat.Bsize))

	// "if file system size is zero, cannot calculate usage percentage"
	if device.FSSize != 0 {
		device.FSUsePercent = (device.FSUsed / device.FSSize) * 100
	}
}

func checkIsPartition(devicePath string) bool {
	ueventPath := filepath.Join(devicePath, "uevent")
	if file, err := os.Open(ueventPath); err == nil { //nolint:gosec
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "DEVTYPE=") {
				if strings.TrimPrefix(line, "DEVTYPE=") == "partition" {
					return true
				}
			}
		}
	}
	return false
}

func setType(device *BlockDeviceStat, devicePath string) {
	if checkIsPartition(devicePath) {
		device.Type = PARTITION
		return
	}
	baseName := filepath.Base(device.Name)

	switch {
	case device.IsDM:
		/* The DM_UUID prefix should be set to subsystem owning
		 * the device - LVM, CRYPT, DMRAID, MPATH, PART */
		if dmUUID := ulPathReadString(devicePath, "dm/uuid"); dmUUID != "" {
			parts := strings.Split(dmUUID, "-")
			if len(parts) >= 2 {
				dmUUIDPrefix := parts[0]
				device.Type = strings.ToLower(dmUUIDPrefix)
			}
		} else {
			device.Type = "dm"
		}
	case strings.HasPrefix(baseName, "loop"):
		device.Type = "loop"
	case strings.HasPrefix(baseName, "md"):
		mdLevel := ulPathReadString(devicePath, "md/level")
		if mdLevel == "" {
			mdLevel = "md"
		}
		device.Type = mdLevel
	default:
		if blkTypeID, err := strconv.ParseInt(ulPathReadString(devicePath, "device/type"), 10, 64); err == nil {
			blkType := blkdevScsiTypeToName(int(blkTypeID))
			device.Type = blkType
			return
		}
		device.Type = DISK
	}
}

func setOwnerAndGroup(device *BlockDeviceStat, devicePath string) {
	fileInfo, err := os.Stat(devicePath)
	if err != nil {
		return
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}

	uid := int(stat.Uid)
	gid := int(stat.Gid)

	if u, err := user.LookupId(fmt.Sprintf("%d", uid)); err == nil {
		device.Owner = u.Username
	}

	if g, err := user.LookupGroupId(fmt.Sprintf("%d", gid)); err == nil {
		device.Group = g.Name
	}
}

func setDeviceUUID(devices *[]BlockDeviceStat) {
	uuidDir := hostDev("disk/by-uuid")
	entries, err := os.ReadDir(uuidDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		linkPath := filepath.Join(uuidDir, entry.Name())
		target, err := filepath.EvalSymlinks(linkPath)
		if err != nil {
			continue
		}

		for i := range *devices {
			if target == makeName((*devices)[i].KName) {
				(*devices)[i].UUID = entry.Name()
			}
		}
	}
}

func setDeviceLabel(devices *[]BlockDeviceStat) {
	labelDir := hostDev("disk/by-label")
	entries, err := os.ReadDir(labelDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		linkPath := filepath.Join(labelDir, entry.Name())
		target, err := filepath.EvalSymlinks(linkPath)
		if err != nil {
			continue
		}

		for i := range *devices {
			if target == makeName((*devices)[i].KName) {
				(*devices)[i].UUID = entry.Name()
			}
		}
	}
}

func getPartitionRQSize(name string, devices *[]BlockDeviceStat) float64 {
	for _, device := range *devices {
		if device.Name == name {
			return device.RQSize
		}
	}
	return 0
}

func setDeviceRQSize(devices *[]BlockDeviceStat) {
	for i := range *devices {
		device := (*devices)[i]
		if device.Parents != nil && device.Type == PARTITION {
			(*devices)[i].RQSize = getPartitionRQSize(device.Parents[0], devices)
		}
	}
}

func cartesianProduct(list1, list2 []string) [][]string {
	if len(list1) == 0 {
		list1 = []string{""}
	}

	if len(list2) == 0 {
		list2 = []string{""}
	}

	var result [][]string

	for _, item1 := range list1 {
		for _, item2 := range list2 {
			result = append(result, []string{item1, item2})
		}
	}

	return result
}
