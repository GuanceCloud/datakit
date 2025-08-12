// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package lsblk collect host lsblk metrics.
package lsblk

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "lsblk"
	metricName  = inputName
)

var (
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type BlockDeviceStat struct {
	Name         string   `json:"name"`
	KName        string   `json:"kname"`
	Parent       string   `json:"parent"`
	Parents      []string `json:"parents"`
	MajMin       string   `json:"maj_min"`
	FSAvail      float64  `json:"fsavail"`
	FSSize       float64  `json:"fssize"`
	FSType       string   `json:"fstype"`
	FSUsed       float64  `json:"fsused"`
	FSUsePercent float64  `json:"fsused_percent"`
	MountPoint   string   `json:"mountpoint"`
	MountPoints  []string `json:"mountpoints"`
	Label        string   `json:"label"`
	UUID         string   `json:"uuid"`
	Model        string   `json:"model"`
	Serial       string   `json:"serial"`
	BlockSize    int64    `json:"block_size"`
	Size         float64  `json:"size"`
	State        string   `json:"state"`
	RQSize       float64  `json:"rq_size"`
	Type         string   `json:"type"`
	Vendor       string   `json:"vendor"`
	Owner        string   `json:"owner"`
	Group        string   `json:"group"`

	IsMounted bool `json:"ismounted"`
	IsDM      bool `json:"is_dm"`
}

type Input struct {
	Interval time.Duration

	Tags          map[string]string `toml:"tags"`
	ExcludeDevice []string          `toml:"exclude_device"`

	collectCache []*point.Point
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger

	semStop      *cliutils.Sem
	partitionMap map[string]*BlockDeviceStat
	ptsTime      time.Time
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	ipt.ptsTime = ntp.Now()
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
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval)
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

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(ipt.ptsTime.UnixNano()))

	if len(ipt.collectCache) > 0 { // clear
		ipt.collectCache = ipt.collectCache[:0]
	}

	lsblkInfo, err := ipt.collectLsblkInfo()
	if err != nil {
		return fmt.Errorf("error getting lsblk info: %w", err)
	}

	for _, device := range lsblkInfo {
		var kvs point.KVs

		kvs = kvs.AddTag("name", device.Name)
		kvs = kvs.AddTag("kname", device.KName)

		kvs = kvs.AddTag("maj_min", device.MajMin)
		kvs = kvs.AddTag("label", device.Label)
		kvs = kvs.AddTag("uuid", device.UUID)
		kvs = kvs.AddTag("serial", device.Serial)
		kvs = kvs.AddTag("model", device.Model)
		kvs = kvs.AddTag("state", device.State)
		kvs = kvs.AddTag("type", device.Type)
		kvs = kvs.AddTag("vendor", device.Vendor)
		kvs = kvs.AddTag("owner", device.Owner)
		kvs = kvs.AddTag("group", device.Group)
		kvs = kvs.AddTag("parent", device.Parent)

		kvs = kvs.Set("size", device.Size)
		kvs = kvs.Set("rq_size", device.RQSize)
		if device.IsMounted {
			kvs = kvs.Set("fs_size", device.FSSize)
			kvs = kvs.Set("fs_avail", device.FSAvail)
			kvs = kvs.Set("fs_used", device.FSUsed)
			kvs = kvs.Set("fs_used_percent", device.FSUsePercent)

			kvs = kvs.AddTag("is_mounted", "是")
			kvs = kvs.AddTag("mountpoint", device.MountPoint)
		} else {
			kvs = kvs.AddTag("is_mounted", "否")
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs, opts...))
	}

	return nil
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string      { return "host" }
func (*Input) SampleConfig() string { return sampleCfg }
func (*Input) AvailableArchs() []string {
	return []string{
		datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker,
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func defaultInput() *Input {
	ipt := &Input{
		Interval:     time.Second * 10,
		Tags:         make(map[string]string),
		feeder:       dkio.DefaultFeeder(),
		semStop:      cliutils.NewSem(),
		tagger:       datakit.DefaultGlobalTagger(),
		mergedTags:   make(map[string]string),
		partitionMap: make(map[string]*BlockDeviceStat),
	}
	return ipt
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval", Type: doc.TimeDuration, Default: "`10s`", Desc: "Collect interval", DescZh: "采集器重复间隔时长"},
		{FieldName: "ExcludeDevice", Type: doc.List, Example: `/dev/loop0,/dev/loop1`, Desc: "Excluded device prefix. (By default, collect all devices with dev as the prefix)", DescZh: "排除的设备前缀。（默认收集以 dev 为前缀的所有设备）"},
	}

	return doc.SetENVDoc("ENV_INPUT_LSBLK_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_LSBLK_INTERVAL : time.Duration
//	ENV_INPUT_LSBLK_EXCLUDE_DEVICE : []string
func (ipt *Input) ReadEnv(envs map[string]string) {
	if fsList, ok := envs["ENV_INPUT_LSBLK_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}

	if tagsStr, ok := envs["ENV_INPUT_LSBLK_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_LSBLK_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_LSBLK_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_LSBLK_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
