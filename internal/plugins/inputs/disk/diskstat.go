// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/pcommon"
)

func (ipt *Input) filterUsage() (arr []pcommon.FilesystemStats, err error) {
	res, err := pcommon.FilterUsage(ipt.diskStats, ipt.hostRoot)
	if err != nil {
		return nil, err
	}

	for _, x := range res {
		if datakit.StrEFInclude(x.Part.Device, ipt.ExcludeDevice) {
			l.Debugf("exclude device %+#v", x.Part)
			continue
		}

		if ipt.regIgnoreFSTypes != nil && ipt.regIgnoreFSTypes.MatchString(x.Part.Fstype) {
			l.Debugf("ignore fs type %s on %+#v", x.Part.Fstype, x.Part)
			continue
		}

		if ipt.regIgnoreMountpoints != nil && ipt.regIgnoreMountpoints.MatchString(x.Part.Mountpoint) {
			l.Debugf("ignore mountpoint %s on %+#v", x.Part.Mountpoint, x.Part)
			continue
		}

		arr = append(arr, x)
	}
	return
}
