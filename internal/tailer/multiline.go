// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"

func newMultiline(cfg *config) (*multiline.Multiline, error) {
	opts := []multiline.Option{multiline.WithMaxLength(int(cfg.maxMultilineLength))}
	if cfg.autoMultiline {
		return multiline.NewAuto(cfg.extraPatterns, opts...)
	}
	if cfg.multilinePattern == "" {
		return multiline.New(nil, opts...)
	}
	return multiline.New([]string{cfg.multilinePattern}, opts...)
}
