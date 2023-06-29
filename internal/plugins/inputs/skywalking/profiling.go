// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	profileV3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/profile/v3"
)

// TODO:.
func processProfileV3(threadSnapshot *profileV3.ThreadSnapshot) {
	log.Debugf("profile = %+v", threadSnapshot)
}
