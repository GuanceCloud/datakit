// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logtail wraps logging collect functions
package logtail

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/recorder"
)

var (
	logtailCachePath = datakit.JoinToCacheDir("logtail.history")

	l = logger.DefaultSLogger("logtail")
)

func InitDefault() error {
	l = logger.SLogger("logtail")
	if err := recorder.Init(logtailCachePath); err != nil {
		l.Warnf("failed to initialize recorder %s", err)
		return err
	}
	return nil
}
