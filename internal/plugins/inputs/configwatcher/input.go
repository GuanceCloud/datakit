// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package configwatcher monitors files for changes.
package configwatcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Input struct {
	Source      string            `toml:"source"`
	MountPoint  string            `toml:"mount_point"`
	Paths       []string          `toml:"paths"`
	Interval    time.Duration     `toml:"interval"`
	Recursive   bool              `toml:"recursive"`
	MaxDiffSize int64             `toml:"max_diff_size"`
	Tags        map[string]string `toml:"tags"`

	fileWatchers []*fileWatcher
	mergedTags   map[string]string

	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
	semStop *cliutils.Sem
	log     *logger.Logger
}

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{inputs.DefaultEmptyMeasurement}
}

func (*Input) AvailableArchs() []string { return []string{datakit.OSLabelLinux} }

func (*Input) Catalog() string { return "host" }

func (ipt *Input) Run() {
	ipt.log = logger.SLogger(inputName + "/" + ipt.Source)
	ipt.log.Info("start")

	if err := ipt.setup(); err != nil {
		ipt.log.Infof("init failure: %s", err)
		ipt.log.Info("exit")
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.log.Info("configwatcher exit")
			return

		case <-ipt.semStop.Wait():
			ipt.log.Info("configwatcher return")
			return

		case <-tick.C:
			ipt.checkChanges()
		}
	}
}

func (ipt *Input) setup() error {
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")

	if ipt.MountPoint == "" && datakit.Docker && config.IsKubernetes() {
		if v := os.Getenv("HOST_ROOT"); v != "" {
			ipt.MountPoint = v
		}
	}

	for _, path := range ipt.Paths {
		w, err := newFileWatcher(filepath.Join(ipt.MountPoint, path),
			withMaxDiffSize(ipt.MaxDiffSize),
			withRecursive(ipt.Recursive),
		)
		if err != nil {
			return err
		}

		ipt.fileWatchers = append(ipt.fileWatchers, w)
	}

	return nil
}

func (ipt *Input) checkChanges() {
	var pts []*point.Point

	for _, w := range ipt.fileWatchers {
		events, err := w.checkChanges()
		if err != nil {
			ipt.log.Warn(err)
			continue
		}

		for idx, event := range events {
			if event.typ == noChange {
				continue
			}

			t := time.Now()
			if event.newState != nil && !event.newState.modTime.IsZero() {
				t = event.newState.modTime
			} else if event.oldState != nil && !event.oldState.modTime.IsZero() {
				t = event.oldState.modTime
			}

			var kvs point.KVs
			kvs = append(kvs, buildDefaultChangeEventKVs()...)
			kvs = append(kvs, point.NewTags(ipt.mergedTags)...)
			kvs = kvs.Add("df_title", generateTitle(ipt.Source, &events[idx]))
			kvs = kvs.Add("df_message", generateMessage(ipt.Source, &events[idx]))

			if event.diff != "" {
				kvs = kvs.Add("diff", event.diff)
			}

			pts = append(pts, point.NewPoint("event", kvs, point.WithTimestamp(t.UnixNano())))
		}
	}

	if err := ipt.feeder.Feed(
		point.KeyEvent,
		pts,
		dkio.WithSource(dkio.FeedSource(inputName, ipt.Source)),
	); err != nil {
		ipt.log.Warnf("feed failed, err: %s", err)
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Source:      "not-set",
			Interval:    time.Minute * 5,
			Recursive:   true,
			MaxDiffSize: 1024 * 256,
			Tags:        make(map[string]string),
			feeder:      dkio.DefaultFeeder(),
			tagger:      datakit.DefaultGlobalTagger(),
			semStop:     cliutils.NewSem(),
		}
	})
}
