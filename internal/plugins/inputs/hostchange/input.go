// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

// Package hostchange collect host config changes.
package hostchange

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/google/uuid"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName       = "host_change"
	defaultInterval = 1 * time.Minute
	maxInterval     = 30 * time.Minute
	minInterval     = 10 * time.Second

	defaultChangeLanguage = changes.LangEn
)

var (
	_ inputs.Input     = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval  time.Duration        `toml:"interval,omitempty"`
	UserGroup UserGroup            `toml:"user_group"` // User and group change detection configuration
	Crontab   CrontabChecker       `toml:"crontab"`    // Crontab change detection configuration
	File      FileChecker          `toml:"file"`       // File change detection configuration
	Service   ServiceChecker       `toml:"service"`    // Service change detection configuration
	Network   NetworkConfigChecker `toml:"network"`    // Network configuration change detection
	Tags      map[string]string    `toml:"tags"`

	feeder     dkio.Feeder
	mergedTags map[string]string
	semStop    *cliutils.Sem
	tagger     datakit.GlobalTagger

	start      time.Time
	collectors map[string]func() ([]*ChangeItem, error)
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	if err := ipt.setup(); err != nil {
		l.Errorf("setup failed: %s", err.Error())
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		return
	}

	ticker := time.NewTicker(ipt.Interval)
	defer ticker.Stop()

	ipt.start = ntp.Now()
	for {
		ipt.doCollect()
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s return on sem", inputName)
			return
		case tt := <-ticker.C:
			ipt.start = inputs.AlignTime(tt, ipt.start, ipt.Interval)
		}
	}
}

func (ipt *Input) setup() error {
	l.Infof("%s input started", inputName)

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")

	// Load host manifest
	if err := changes.LoadHostManifest(); err != nil {
		return fmt.Errorf("failed to load host manifest: %w", err)
	}

	// Setup collectors
	ipt.setupCollectors()

	return nil
}

func (ipt *Input) setupCollectors() {
	ipt.collectors = make(map[string]func() ([]*ChangeItem, error))

	if ipt.UserGroup.Enabled {
		if err := ipt.UserGroup.Init(ipt); err != nil {
			l.Warnf("user group change init failed: %s", err.Error())
		} else {
			ipt.collectors["user_group"] = func() ([]*ChangeItem, error) {
				return ipt.UserGroup.Collect()
			}
		}
	}

	// Add crontab collector if enabled
	if ipt.Crontab.Enabled {
		if err := ipt.Crontab.Init(ipt); err != nil {
			l.Warnf("crontab change init failed: %s", err.Error())
		} else {
			ipt.collectors["crontab"] = func() ([]*ChangeItem, error) {
				return ipt.Crontab.Collect()
			}
		}
	}

	// Add file collector if enabled
	if ipt.File.Enabled {
		if err := ipt.File.Init(ipt); err != nil {
			l.Warnf("file change init failed: %s", err.Error())
		} else {
			ipt.collectors["file"] = func() ([]*ChangeItem, error) {
				return ipt.File.Collect()
			}
		}
	}

	// Add service collector if enabled
	if ipt.Service.Enabled {
		if err := ipt.Service.Init(ipt); err != nil {
			l.Warnf("service change init failed: %s", err.Error())
		} else {
			ipt.collectors["service"] = func() ([]*ChangeItem, error) {
				return ipt.Service.Collect()
			}
		}
	}

	// Add network collector if enabled
	if ipt.Network.Enabled {
		if err := ipt.Network.Init(ipt); err != nil {
			l.Warnf("network change init failed: %s", err.Error())
		} else {
			ipt.collectors["network"] = func() ([]*ChangeItem, error) {
				return ipt.Network.Collect()
			}
		}
	}
}

type ChangeItem struct {
	ChangeID             changes.ChangeID
	ChangeTimestampMicro int64 // Change timestamp in microseconds
	Title,
	Message string
}

func (ipt *Input) feedPointFromChangeItems(changes []*ChangeItem) error {
	if len(changes) == 0 {
		return nil
	}

	var pts []*point.Point

	for _, change := range changes {
		var kvs point.KVs

		kvs = append(kvs, buildDefaultChangeEventKVs()...)
		kvs = kvs.AddTag("change_id", string(change.ChangeID))
		changeTime := time.Now().UnixMicro()
		if change.ChangeTimestampMicro > 0 {
			changeTime = change.ChangeTimestampMicro
		}

		kvs = kvs.Add("change_time_us", changeTime)
		kvs = kvs.Add("df_title", change.Title)
		kvs = kvs.Add("df_message", change.Message)

		kvs = append(kvs, point.NewTags(ipt.mergedTags)...)
		pts = append(pts, point.NewPoint("event", kvs, point.WithTimestamp(ipt.start.UnixNano())))
	}

	if err := ipt.feeder.Feed(
		point.KeyEvent,
		pts,
		dkio.WithSource("host-change-event"),
	); err != nil {
		l.Warnf("feed failed, err: %s", err)
		return fmt.Errorf("feed failed: %w", err)
	}

	return nil
}

func (ipt *Input) doCollect() {
	for collectorName, collector := range ipt.collectors {
		l.Debugf("start to collect %s change", collectorName)
		changeItems, err := collector()
		if err != nil {
			l.Errorf("collect %s failed, err: %s", collectorName, err)
			continue
		}

		if err := ipt.feedPointFromChangeItems(changeItems); err != nil {
			l.Errorf("feed %s failed, err: %s", collectorName, err)
			continue
		}
	}
}

func buildDefaultChangeEventKVs() (kvs point.KVs) {
	const (
		defaultStatus = "info"
		defaultSource = "change"
	)

	var uid string
	if u, err := uuid.NewRandom(); err == nil {
		uid = "event-" + strings.ToLower(u.String())
	} else {
		l.Warnf("cannot generate UUIDv4, err: %s", err)
	}
	kvs = kvs.AddTag("df_event_id", uid)
	kvs = kvs.AddTag("df_source", defaultSource)
	kvs = kvs.AddTag("df_status", defaultStatus)
	kvs = kvs.AddTag("df_sub_status", defaultStatus)

	return
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return []string{datakit.OSLabelLinux} }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ChangeMeasurement{},
	}
}

func defaultInput() *Input {
	return &Input{
		Interval: defaultInterval,
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
		Tags:     make(map[string]string),
		UserGroup: UserGroup{
			Enabled: true, // Enable users and groups detection by default
		},
		Crontab: CrontabChecker{
			Enabled: true, // Enable crontab detection by default
		},
		File: FileChecker{
			Enabled: false,
		},
		Service: ServiceChecker{
			Enabled:      true,                // Enable service detection by default
			ServiceTypes: []string{"systemd"}, // Default to systemd
		},
		Network: NetworkConfigChecker{
			Enabled: true,
		},
		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
