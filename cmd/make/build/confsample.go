// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func CheckSampleConf(allInputs map[string]inputs.Creator) error {
	l.Infof("check %d inputs's config samples", len(allInputs))

	for name, creator := range allInputs {
		l.Debugf("check config sample of %q", name)
		input := creator()

		// check if all conf sample are EN
		if err := isAllASCII(input.SampleConfig()); err != nil {
			return fmt.Errorf("got non-ASCII characters for input %q's sample.conf: %w", name, err)
		}
	}

	// check datakit.conf.sample
	if err := isAllASCII(datakit.MainConfSample(datakit.BrandDomainTemplate)); err != nil {
		return fmt.Errorf("got non-ASCII characters within datakit.conf sample: %w", err)
	}

	return nil
}

func isAllASCII(s string) error {
	var line int
	for _, ch := range s {
		if ch == '\n' {
			line++
		}

		if ch > 127 {
			return fmt.Errorf("get non-ascii character %q at line %d", ch, line)
		}
	}

	return nil
}
