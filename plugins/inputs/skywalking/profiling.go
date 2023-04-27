// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	profileV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/language/profile/v3"
)

// TODO:.
func processProfileV3(threadSnapshot *profileV3.ThreadSnapshot) {
	log.Debugf("profile = %+v", threadSnapshot)
}
