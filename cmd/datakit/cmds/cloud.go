// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject"
)

func showCloudInfo() (map[string]interface{}, error) {
	hostobject.SetLog()

	x := hostobject.Input{Tags: map[string]string{}}
	if err := x.SetCloudProvider(); err != nil {
		return nil, fmt.Errorf("failed to decide cloud provider: %w", err)
	}
	return x.SyncCloudInfo(x.Tags["cloud_provider"])
}
