// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hostdir collect directory metrics.
package hostdir

import (
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) SampleConfig() string {
	return sample
}

func (ipt *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, "")

	pt := point.NewPointV2(name,
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)

	ipt.collectCache = append(ipt.collectCache, pt)
}

func (*Input) Catalog() string {
	return "host"
}

func (ipt *Input) collect() error {
	timeNow := time.Now()
	var tags map[string]string
	path := ipt.Dir
	if ipt.platform == datakit.OSWindows {
		filesystem, err := GetFileSystemType(path)
		if err != nil {
			return err
		}
		dirMode, err := Getdirmode(path)
		if err != nil {
			return err
		}
		tags = map[string]string{
			"file_mode":      dirMode,
			"file_system":    filesystem,
			"host_directory": ipt.Dir,
		}
	} else {
		filesystem, err := GetFileSystemType(path)
		if err != nil {
			return err
		}
		dirMode, err := Getdirmode(path)
		if err != nil {
			return err
		}
		fileownership, err := GetFileOwnership(path, ipt.platform)
		if err != nil {
			fileownership = "N/A"
		}
		tags = map[string]string{
			"file_mode":      dirMode,
			"file_system":    filesystem,
			"file_ownership": fileownership,
			"host_directory": ipt.Dir,
		}
	}

	for k, v := range ipt.Tags {
		tags[k] = v
	}
	filesize, filecount, dircount := Startcollect(ipt.Dir, ipt.ExcludePatterns)
	fields := map[string]interface{}{
		"file_size":  filesize,
		"file_count": filecount,
		"dir_count":  dircount,
	}
	ipt.appendMeasurement(inputName, tags, fields, timeNow)
	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("hostdir input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		ipt.collectCache = make([]*point.Point, 0)

		start := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect failed: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(metricName,
				point.Metric,
				ipt.collectCache,
				&io.Option{CollectCost: time.Since((start))}); err != nil {
				l.Errorf("Feed failed: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("hostdir input exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof("hostdir input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},
			platform: runtime.GOOS,

			semStop: cliutils.NewSem(),
			feeder:  io.DefaultFeeder(),
			Tagger:  datakit.DefaultGlobalTagger(),
		}
		return s
	})
}
