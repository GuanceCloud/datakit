// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing, metrics and logging.
package skywalking

import profileV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/profile/v3"

func processProfile(threadSnapshot *profileV3.ThreadSnapshot) {
	// todo
	log.Debugf("profile = %+v", threadSnapshot)
}
