// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package disk collect host disk metrics.
package disk

import (
	"os"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			FieldName: "Interval",
		},
		{
			FieldName: "Tags",
		},
		{
			FieldName: "ExtraDevice",
			Type:      doc.List,
			Example:   "`/nfsdata,other_data`",
			Desc:      "Additional device prefix. (By default collect all devices with dev as the prefix)",
			DescZh:    "额外的设备前缀。（默认收集以 dev 为前缀的所有设备）",
		},
		{
			FieldName: "ExcludeDevice",
			Type:      doc.List,
			Example:   `/dev/loop0,/dev/loop1`,
			Desc:      "Excluded device prefix. (By default collect all devices with dev as the prefix)",
			DescZh:    "排除的设备前缀。（默认收集以 dev 为前缀的所有设备）",
		},

		{
			FieldName: "IgnoreMountpoints",
			Type:      doc.String,
			Example:   "`^(/usr/local/datakit/.*|/run/containerd/.*)$`",
			Desc:      "Excluded mount points",
			DescZh:    "忽略这些挂载点对应的磁盘指标",
		},

		{
			FieldName: "IgnoreFSTypes",
			Type:      doc.String,
			ENVName:   "INPUT_DISK_IGNORE_FSTYPES",
			Example:   "`^(tmpfs|autofs|binfmt_misc|devpts|fuse.lxcfs|overlay|proc|squashfs|sysfs)$`",
			Desc:      "Excluded file systems",
			DescZh:    "忽略这些文件系统对应的磁盘指标",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_DISK_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_DISK_EXCLUDE_DEVICE : []string
//	ENV_INPUT_DISK_EXTRA_DEVICE : []string
//	ENV_INPUT_DISK_TAGS : "a=b,c=d"
//	ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE : bool
//	ENV_INPUT_DISK_INTERVAL : time.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if fsList, ok := envs["ENV_INPUT_DISK_EXTRA_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add extra_device from ENV: %v", fsList)
		ipt.ExtraDevice = append(ipt.ExtraDevice, list...)
	}
	if fsList, ok := envs["ENV_INPUT_DISK_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}

	if tagsStr, ok := envs["ENV_INPUT_DISK_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	// ENV_INPUT_DISK_INTERVAL : time.Duration
	// ENV_INPUT_DISK_MOUNT_POINTS : []string
	if str, ok := envs["ENV_INPUT_DISK_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISK_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str := envs["ENV_INPUT_DISK_IGNORE_FSTYPES"]; str != "" {
		ipt.IgnoreFSTypes = str
	}

	if str := envs["ENV_INPUT_DISK_IGNORE_MOUNTPOINTS"]; str != "" {
		ipt.IgnoreMountpoints = str
	}

	// Default setting: we have add the env HOST_ROOT in datakit.yaml by default
	// but some old deployments may not hava this ENV set.
	ipt.hostRoot = "/rootfs"

	// Deprecated: use ENV_HOST_ROOT
	if v := os.Getenv("HOST_ROOT"); v != "" {
		ipt.hostRoot = v
	}

	if v := os.Getenv("ENV_HOST_ROOT"); v != "" {
		ipt.hostRoot = v
	}
}
