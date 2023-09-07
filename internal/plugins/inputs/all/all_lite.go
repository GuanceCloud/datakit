// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build datakit_lite && with_inputs
// +build datakit_lite,with_inputs

// Package inputs wraps all inputs implements
package inputs

import (
	// default inputs.
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/cpu"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/dialtesting"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/disk"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/diskio"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/dk"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/hostobject"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/logging"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/mem"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/net"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/process"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/swap"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/system"
)
