// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject"
)

func showCloudInfo(cloud string) (map[string]interface{}, error) {
	hostobject.SetLog()

	x := hostobject.Input{}
	return x.SyncCloudInfo(cloud)
}
