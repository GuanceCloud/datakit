// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package nfs provides functionality related to NFS (Network File System) monitoring and metrics collection.
package nfs

import (
	"runtime"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval time.Duration

	Tags             map[string]string `toml:"tags"`
	MountstatsMetric mountstatsMetric  `toml:"mountstats"`
	NFSd             bool              `toml:"nfsd"`

	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger
	collectors   []func() ([]*point.Point, error)
	collectCache []*point.Point

	semStop *cliutils.Sem
	alignTS int64
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	lastTS := time.Now()
	for {
		ipt.alignTS = lastTS.UnixNano()

		start := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(false),
				dkio.WithInputName(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, lastTS.UnixMilli(), ipt.Interval.Milliseconds())
			lastTS = time.UnixMilli(nextts)
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) collect() error {
	if runtime.GOOS != datakit.OSLinux {
		l.Infof("collect nfs not implemented under %s", runtime.GOOS)
		return nil
	}

	ipt.collectCache = make([]*point.Point, 0)

	if len(ipt.collectors) == 0 {
		ipt.collectors = []func() ([]*point.Point, error){
			ipt.collectMountStats,
			ipt.collectBase,
		}
	}

	var ptsMetric []*point.Point
	for idx, f := range ipt.collectors {
		l.Debugf("collecting %d(%v)...", idx, f)

		pts, err := f()
		if err != nil {
			l.Errorf("collect failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if ipt.NFSd {
		pts, err := ipt.collectNFSd()
		if err != nil {
			l.Errorf("collect NFSd failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	ipt.collectCache = ptsMetric
	return nil
}

func (ipt *Input) collectMountStats() ([]*point.Point, error) {
	pts, err := ipt.buildMountStats()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

func (ipt *Input) collectBase() ([]*point.Point, error) {
	pts, err := ipt.buildBaseMetric()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

func (ipt *Input) collectNFSd() ([]*point.Point, error) {
	pts, err := ipt.buildNFSdMetric()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
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
		&baseMeasurement{},
		&mountstatsMeasurement{},
		&nfsdMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval", Type: doc.TimeDuration, Default: "`10s`", Desc: "Collect interval", DescZh: "采集器重复间隔时长"},
		{
			FieldName: "EnableMountStatsRwBytes",
			Type:      doc.Boolean,
			Default:   "`false`",
			Desc:      "Enable detailed read and write bytes information for NFS mount points",
			DescZh:    "开启 NFS 挂载点的详细字节读写信息",
		},
		{FieldName: "EnableMountStatsTransport", Type: doc.Boolean, Default: "`false`", Desc: "Enable NFS mount point and server transfer information", DescZh: "开启 NFS 挂载点与服务端的传输信息"},
		{FieldName: "EnableMountStatsEvent", Type: doc.Boolean, Default: "`false`", Desc: "Event statistics", DescZh: "开启 NFS 事件统计信息"},
		{FieldName: "EnableMountStatsOperations", Type: doc.Boolean, Default: "`false`", Desc: "Enable NFS transfer information for a given operation", DescZh: "开启 NFS 给定操作的传输信息"},
		{
			FieldName: "nfsd",
			Type:      doc.Boolean,
			Default:   "`false`",
			Desc:      "Enable the NFSd indicator",
			DescZh:    "开启 NFSd 指标",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_NFS_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_NFS_INTERVAL : time.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_NFS_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_NFS_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_NFS_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_NFS_ENABLE_MOUNTSTATS_RW_BYTES"]; ok {
		flag, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_ENABLE_MOUNTSTATS_RW_BYTES: %s, ignore", err)
		} else {
			ipt.MountstatsMetric.Rw = flag
		}
	}

	if str, ok := envs["ENV_INPUT_NFS_ENABLE_MOUNTSTATS_TRANSPORT"]; ok {
		flag, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_ENABLE_MOUNTSTATS_TRANSPORT: %s, ignore", err)
		} else {
			ipt.MountstatsMetric.Transport = flag
		}
	}

	if str, ok := envs["ENV_INPUT_NFS_ENABLE_MOUNTSTATS_EVENT"]; ok {
		flag, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_ENABLE_MOUNTSTATS_EVENT: %s, ignore", err)
		} else {
			ipt.MountstatsMetric.Event = flag
		}
	}

	if str, ok := envs["ENV_INPUT_NFS_ENABLE_MOUNTSTATS_OPERATIONS"]; ok {
		flag, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_ENABLE_MOUNTSTATS_OPERATIONS: %s, ignore", err)
		} else {
			ipt.MountstatsMetric.Operations = flag
		}
	}

	if str, ok := envs["ENV_INPUT_NFS_NFSD"]; ok {
		flag, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NFS_NFSD: %s, ignore", err)
		} else {
			ipt.NFSd = flag
		}
	}
}

func defaultInput() *Input {
	ipt := &Input{
		Interval:   time.Second * 10,
		NFSd:       false,
		Tags:       make(map[string]string),
		feeder:     dkio.DefaultFeeder(),
		semStop:    cliutils.NewSem(),
		tagger:     datakit.DefaultGlobalTagger(),
		mergedTags: make(map[string]string),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
