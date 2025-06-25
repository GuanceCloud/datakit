// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package xfsquota implements the collection of quota information for the XFS file system.
package xfsquota

import (
	"fmt"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Input struct {
	BinaryPath     string            `toml:"binary_path"`
	Interval       time.Duration     `toml:"interval"`
	FilesystemPath string            `toml:"filesystem_path"`
	ParserVersion  string            `toml:"parser_version"`
	Tags           map[string]string `toml:"tags"`

	Feeder dkio.Feeder
	Tagger datakit.GlobalTagger
}

var l = logger.DefaultSLogger(inputName)

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&xfsquotaMetric{}}
}

func (*Input) AvailableArchs() []string { return []string{datakit.OSLabelLinux} }

func (*Input) Catalog() string { return "xfsquota" }

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("xfsquota start")
	if err := ipt.setup(); err != nil {
		l.Warn(err)
		return
	}

	start := ntp.Now()
	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		ipt.collectXFSQuota(start.UnixNano())

		select {
		case <-datakit.Exit.Wait():
			l.Info("xfsquota exit")
			return

		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)
		}
	}
}

func (ipt *Input) setup() error {
	if _, err := os.Stat(ipt.BinaryPath); err != nil {
		return fmt.Errorf("invalid binary path, err: %w", err)
	}
	if stat, err := os.Stat(ipt.FilesystemPath); err != nil {
		return fmt.Errorf("invalid filesystem path, err: %w", err)
	} else if !stat.IsDir() {
		return fmt.Errorf("filesystem path %s is not a directory", ipt.FilesystemPath)
	}

	return nil
}

func (ipt *Input) collectXFSQuota(timestamp int64) {
	start := time.Now()

	quotaInfo, err := getXFSQuota(ipt.BinaryPath, ipt.FilesystemPath)
	if err != nil {
		l.Warn("exec failed: %s", err)
		return
	}
	quotaList, err := parseQuotaOutput(quotaInfo)
	if err != nil {
		l.Warn("parse failed: %s", err)
		return
	}

	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, quota := range quotaList {
		var kvs point.KVs
		kvs = kvs.AddTag("project_id", quota.ProjectID)
		kvs = kvs.AddTag("filesystem_path", ipt.FilesystemPath)

		kvs = kvs.Add("used", quota.UsedBlocks, false, true)
		kvs = kvs.Add("soft", quota.SoftLimit, false, true)
		kvs = kvs.Add("hard", quota.HardLimit, false, true)

		for key, value := range ipt.Tags {
			kvs = kvs.AddTag(key, value)
		}
		pts = append(pts, point.NewPointV2("xfsquota", kvs, append(opts, point.WithTimestamp(timestamp))...))
	}

	if err := ipt.Feeder.Feed(
		point.Metric,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithSource(inputName),
	); err != nil {
		l.Warnf("failed to feed xfsquota metrics: %s", err)
	}
}

func (ipt *Input) Terminate() { /*nil*/ }

func newXFSQuota() *Input {
	return &Input{
		Tags:   make(map[string]string),
		Feeder: dkio.DefaultFeeder(),
		Tagger: datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newXFSQuota()
	})
}
