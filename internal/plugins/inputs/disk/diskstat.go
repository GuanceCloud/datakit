// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/pcommon"
)

func (ipt *Input) filterUsage() (arr []pcommon.FilesystemStats, err error) {
	parts, err := ipt.diskStats.Partitions()
	if err != nil {
		return nil, fmt.Errorf("Partitions(): %w", err)
	}

	for i := range parts {
		part := &parts[i]
		trimmedPart := pcommon.TrimPartitionHostPath(ipt.hostRoot, part)

		if datakit.StrEFInclude(trimmedPart.Device, ipt.ExcludeDevice) {
			l.Debugf("exclude device %+#v", trimmedPart)
			continue
		}

		if ipt.regIgnoreFSTypes != nil && ipt.regIgnoreFSTypes.MatchString(trimmedPart.Fstype) {
			l.Debugf("ignore fs type %s on %+#v", trimmedPart.Fstype, trimmedPart)
			continue
		}

		if ipt.regIgnoreMountpoints != nil && ipt.regIgnoreMountpoints.MatchString(trimmedPart.Mountpoint) {
			l.Debugf("ignore mountpoint %s on %+#v", trimmedPart.Mountpoint, trimmedPart)
			continue
		}

		du, err := ipt.diskStats.Usage(part.Mountpoint, ipt.hostRoot)
		if err != nil {
			l.Warnf("Usage on partition %+#v: %s, ignored", part, err)
			continue
		}

		du.Path = trimmedPart.Mountpoint
		arr = append(arr, pcommon.FilesystemStats{
			Usage: du,
			Part:  trimmedPart,
		})
	}
	return arr, nil
}
