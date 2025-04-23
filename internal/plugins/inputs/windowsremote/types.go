// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// nolint
type (
	HostMetaInfo struct {
		HostName        string `json:"host_name"`
		BootTime        uint64 `json:"boot_time"`
		OS              string `json:"os"`
		Platform        string `json:"platform"`         // ex: ubuntu, linuxmint
		PlatformFamily  string `json:"platform_family"`  // ex: debian, rhel
		PlatformVersion string `json:"platform_version"` // version of the complete OS
		Kernel          string `json:"kernel_release"`
		Arch            string `json:"arch"`
	}

	CPUInfo struct {
		Vendor    string  `json:"vendor_id"`
		Module    string  `json:"module_name"`
		Cores     int32   `json:"cores"`
		Mhz       float64 `json:"mhz"`
		CacheSize int32   `json:"cache_size"`
		Flags     string  `json:"-"`
	}

	MemInfo struct {
		MemoryTotal uint64 `json:"memory_total"`
		SwapTotal   uint64 `json:"swap_total"`
		usedPercent float64
	}

	NetInfo struct {
		Index        int      `json:"-"`
		MTU          int      `json:"mtu"`   // maximum transmission unit
		Name         string   `json:"name"`  // e.g., "en0", "lo0", "eth0.100"
		HardwareAddr string   `json:"mac"`   // IEEE MAC-48, EUI-48 and EUI-64 form
		Flags        []string `json:"flags"` // e.g., FlagUp, FlagLoopback, FlagMulticast
		IP4          string   `json:"ip4"`
		IP6          string   `json:"ip6"`
		IP4All       []string `json:"ip4_all"`
		IP6All       []string `json:"ip6_all"`
		Addrs        []string `json:"-"`
	}

	DiskInfo struct {
		Device     string `json:"device"`
		Total      uint64 `json:"total"`
		Fstype     string `json:"fstype"`
		MountPoint string `json:"mountpoint"`
		Opts       string `json:"-"`
	}

	HostInfo struct {
		HostMeta *HostMetaInfo `json:"meta"`
		CPU      []*CPUInfo    `json:"cpu"`
		Mem      *MemInfo      `json:"mem"`
		Net      []*NetInfo    `json:"net"`
		Disk     []*DiskInfo   `json:"disk"`
		// Conntrack              *conntrackutil.Info    `json:"conntrack"`
		// FileFd                 *filefdutil.Info       `json:"filefd"`
		// Election               *election.ElectionInfo `json:"election"`
		ConfigFile             map[string]string `json:"config_file"`
		cpuPercent             float64
		load5                  float64
		cloudInfo              map[string]interface{}
		diskUsedPercent        float64
		diskIOReadBytesPerSec  int64
		diskIOWriteBytesPerSec int64
		netRecvBytesPerSec     int64
		netSendBytesPerSec     int64
		loggingLevel           string
	}
	HostConfig struct {
		IP         string            `json:"ip"`
		DCAConfig  *config.DCAConfig `json:"dca_config"`
		HTTPListen string            `json:"http_listen"`
	}

	HostObjectMessage struct {
		Host       *HostInfo             `json:"host"`
		Collectors []*io.CollectorStatus `json:"collectors,omitempty"`
		Config     *HostConfig           `json:"config"`
	}

	HostprocessesObjectMessage struct {
		Pid         int64  `json:"pid"`
		Name        string `json:"name"`
		ProcessName string `json:"process_name"`
		Cmdline     string `json:"cmdline"`
	}

	CollectorStatus struct {
		Name        string `json:"name"`
		Count       int64  `json:"count"`
		Version     string `json:"version,omitempty"`
		LastTime    int64  `json:"last_time,omitempty"`
		LastErr     string `json:"last_err,omitempty"`
		LastErrTime int64  `json:"last_err_time,omitempty"`
	}
)
