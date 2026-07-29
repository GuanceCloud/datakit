// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package register enables the standalone netpath input.
package register

import netpathinput "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netpath"

func init() { //nolint:gochecknoinits
	netpathinput.Register()
}
