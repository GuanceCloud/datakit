// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hostdir collect directory metrics.
package hostdir

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	dktime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "hostdir"
	metricName  = inputName
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Dir             string        `toml:"dir"`
	ExcludePatterns []string      `toml:"exclude_patterns"`
	Interval        time.Duration `toml:"interval"`
	collectCache    []*point.Point
	Tags            map[string]string `toml:"tags"`
	platform        string

	regSlice   []*regexp.Regexp
	start      time.Time
	semStop    *cliutils.Sem // start stop signal
	feeder     dkio.Feeder
	mergedTags map[string]string
	tagger     datakit.GlobalTagger
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := dktime.NewAlignedTicker(ipt.Interval)
	defer tick.Stop()
	ipt.start = ntp.Now()

	l.Infof("%s input started at timestamp: %d", inputName, time.Now().UnixNano())

	for {
		collectStart := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithElection(false),
				dkio.WithSource(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			ipt.start = inputs.AlignTime(tt, ipt.start, ipt.Interval)
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	for _, v := range ipt.ExcludePatterns {
		reg, err := regexp.Compile(`^.+\.` + v + `$`)
		if err != nil {
			l.Errorf("error regexp: %s", `^.+\.`+v+`$`)
		} else {
			ipt.regSlice = append(ipt.regSlice, reg)
		}
	}
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)

	var (
		filesize  int64
		filecount int64
		dircount  int64
	)
	err := filepath.WalkDir(ipt.Dir, func(name string, di fs.DirEntry, err error) error {
		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("input is exit")
		case <-ipt.semStop.Wait():
			return fmt.Errorf("input is return")
		default:
		}

		if err != nil {
			return err
		}

		info, err := di.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			dircount++
		} else {
			for _, reg := range ipt.regSlice {
				if len(reg.FindAllStringSubmatch(name, 1)) != 0 {
					return nil
				}
			}
			filecount++
		}
		filesize += info.Size()

		return nil
	})
	if err != nil {
		l.Errorf("walk dir %s", err)
		return err
	}

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ipt.start.UnixNano()))
	var kvs point.KVs

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
		kvs = kvs.SetTag("file_mode", dirMode)
		kvs = kvs.SetTag("file_system", filesystem)
		kvs = kvs.SetTag("host_directory", ipt.Dir)
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
		kvs = kvs.SetTag("file_mode", dirMode)
		kvs = kvs.SetTag("file_system", filesystem)
		kvs = kvs.SetTag("file_ownership", fileownership)
		kvs = kvs.SetTag("host_directory", ipt.Dir)
	}

	kvs = kvs.Set("file_size", filesize)
	kvs = kvs.Set("file_count", filecount+dircount)
	kvs = kvs.Set("dir_count", dircount)

	if err = getFileSystemInfo(path, filesize, filecount+dircount, &kvs); err != nil {
		return err
	}

	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs, opts...))

	return nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: time.Second * 10,
			platform: runtime.GOOS,

			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			tagger:  datakit.DefaultGlobalTagger(),
			Tags:    make(map[string]string),
		}
	})
}
